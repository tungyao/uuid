package uuid

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type IdWorker struct {
	startTime             int64
	workerIdBits          uint
	datacenterIdBits      uint
	maxWorkerId           int64
	maxDatacenterId       int64
	sequenceBits          uint
	workerIdLeftShift     uint
	datacenterIdLeftShift uint
	timestampLeftShift    uint
	sequenceMask          int64
	workerId              int64
	datacenterId          int64
	sequence              int64
	lastTimestamp         int64
	signMask              int64
	idLock                *sync.Mutex
}

func (w *IdWorker) InitIdWorker(workerId, datacenterId int64) error {
	var baseValue int64 = -1
	w.startTime = 1603702714272
	w.workerIdBits = 4
	w.datacenterIdBits = 4
	w.maxWorkerId = baseValue ^ (baseValue << w.workerIdBits)
	w.maxDatacenterId = baseValue ^ (baseValue << w.datacenterIdBits)
	w.sequenceBits = 16
	w.workerIdLeftShift = w.sequenceBits
	w.datacenterIdLeftShift = w.workerIdBits + w.workerIdLeftShift
	w.timestampLeftShift = w.datacenterIdBits + w.datacenterIdLeftShift
	w.sequenceMask = baseValue ^ (baseValue << w.sequenceBits)
	w.sequence = 0
	w.lastTimestamp = -1
	w.signMask = ^baseValue + 1

	w.idLock = &sync.Mutex{}

	if w.workerId < 0 || w.workerId > w.maxWorkerId {
		return errors.New(fmt.Sprintf("workerId[%v] is less than 0 or greater than maxWorkerId[%v].", workerId, datacenterId))
	}
	if w.datacenterId < 0 || w.datacenterId > w.maxDatacenterId {
		return errors.New(fmt.Sprintf("datacenterId[%d] is less than 0 or greater than maxDatacenterId[%d].", workerId, datacenterId))
	}
	w.workerId = workerId
	w.datacenterId = datacenterId
	return nil
}

func (w *IdWorker) NextId() int64 {
	w.idLock.Lock()
	timestamp := time.Now().UnixNano()
	if timestamp < w.lastTimestamp {
		panic(errors.New(fmt.Sprintf("Clock moved backwards.  Refusing to generate id for %d milliseconds", w.lastTimestamp-timestamp)))
	}

	if timestamp == w.lastTimestamp {
		w.sequence = (w.sequence + 1) & w.sequenceMask
		if w.sequence == 0 {
			timestamp = w.tilNextMillis()
			w.sequence = 0
		}
	} else {
		w.sequence = 0
	}

	w.lastTimestamp = timestamp

	w.idLock.Unlock()

	id := ((timestamp - w.startTime) << w.timestampLeftShift) |
		(w.datacenterId << w.datacenterIdLeftShift) |
		(w.workerId << w.workerIdLeftShift) |
		w.sequence

	if id < 0 {
		id = -id
	}
	return id
}
func (w *IdWorker) NextStr() string {
	return strconv.Itoa(int(w.NextId()))
}
func (w *IdWorker) NextShort() string {
	return decimalToAny(w.NextId(), 62)
}
func (w *IdWorker) tilNextMillis() int64 {
	timestamp := time.Now().UnixNano()
	if timestamp <= w.lastTimestamp {
		timestamp = time.Now().UnixNano() / int64(time.Millisecond)
	}
	return timestamp
}

var tenToAny = map[int64]string{
	0:  "0",
	1:  "1",
	2:  "2",
	3:  "3",
	4:  "4",
	5:  "5",
	6:  "6",
	7:  "7",
	8:  "8",
	9:  "9",
	10: "a",
	11: "b",
	12: "c",
	13: "d",
	14: "e",
	15: "f",
	16: "g",
	17: "h",
	18: "i",
	19: "j",
	20: "k",
	21: "l",
	22: "m",
	23: "n",
	24: "o",
	25: "p",
	26: "q",
	27: "r",
	28: "s",
	29: "t",
	30: "u",
	31: "v",
	32: "w",
	33: "x",
	34: "y",
	35: "z",
	36: "A",
	37: "B",
	38: "C",
	39: "D",
	40: "E",
	41: "F",
	42: "G",
	43: "H",
	44: "I",
	45: "J",
	46: "K",
	47: "L",
	48: "M",
	49: "N",
	50: "O",
	51: "P",
	52: "Q",
	53: "R",
	54: "S",
	55: "T",
	56: "U",
	57: "V",
	58: "W",
	59: "X",
	60: "Y",
	61: "Z",
}

func decimalToAny(num int64, n int64) string {
	newNumStr := ""
	var remainder int64
	var remainderString string
	for num != 0 {
		remainder = num % n
		remainderString = tenToAny[remainder]
		newNumStr = remainderString + newNumStr
		num = num / n
	}
	return newNumStr
}

func findKey(in string) int {
	var result int = -1
	for k, v := range tenToAny {
		if in == v {
			result = int(k)
		}
	}
	return result
}

func anyToDecimal(num string, n int) string {
	var newNum int
	newNum = 0
	nNum := len(strings.Split(num, "")) - 1
	for _, value := range strings.Split(num, "") {
		tmp := findKey(value)
		if tmp != -1 {
			newNum = newNum + tmp*int(math.Pow(float64(n), float64(nNum)))
			nNum = nNum - 1
		} else {
			break
		}
	}
	return strconv.Itoa(newNum)
}
