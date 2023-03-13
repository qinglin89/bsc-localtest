package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/qinglin89/gobsc/agent"
)

const valCurrent = "0x3b282587d4d4b197117a4279abe886c905538cc0"

const lenV = 5

type recents struct {
	list []string
	idx  int
	l    int
}

func (r *recents) Push(s string) {
	r.list[r.idx] = s
	r.idx = (r.idx + 1) % r.l
}

func (r *recents) CheckRecent(val string) bool {
	for _, v := range r.list {
		if val == v {
			return true
		}
	}
	return false
}

func NewRecents(length int) *recents {
	return &recents{
		list: make([]string, length),
		idx:  0,
		l:    length,
	}
}

type vals [lenV]string

func (v *vals) GetPrev(s string) (string, int) {
	x := v[0]
	if x == s {
		return v[len(v)-1], len(v) - 1
	}
	for i := 1; i < len(v); i++ {
		if v[i] == s {
			return v[i-1], i - 1
		}
	}
	return "", -1
}

func (v *vals) Update(i int, s string) {
	v[i] = s
}

func (v *vals) CheckRecent(val string, inTurn int) ([]string, bool) {
	l := len(v) / 2
	lv := len(v)
	flag := false
	re := make([]string, l)
	for i := 0; i < l; i++ {
		vTmp := v[(inTurn+lv-l+i)%lv]
		re = append(re, vTmp)
		if val == vTmp {
			flag = true
		}
	}
	return re, flag
}

func (v *vals) CalBlockNumberBySlash(nSlashed int64, offTurnVal string, idxInTurn int) int64 {
	nSlashed++
	idxInTurn++
	for offTurnVal != v[idxInTurn%len(v)] {
		nSlashed++
		idxInTurn++
	}
	return nSlashed
}

func main() {
	defer agent.DefaultBsc.StartMiner()
	for {
		fmt.Println("\n===============================================")
		fmt.Println("WAITING...")
		validators, n := GetValidators()
		fmt.Println(*validators)
		prev, idxPrev := validators.GetPrev(valCurrent)
		idxCur := idxPrev + 1
		if prev == "" {
			panic("validator not exists")
		}
		nPrev, validators := WaitForVal(n, prev, validators)
		fmt.Println("SLASH: start on blockNumber:", nPrev+1)
		agent.DefaultBsc.StopMiner()
		n, _ = agent.DefaultBsc.GetBlockNumber()
		for n == nPrev {
			time.Sleep(1 * time.Second)
			n, _ = agent.DefaultBsc.GetBlockNumber()
		}
		if n > nPrev+1 {
			fmt.Println("RESTART LOOP: something wrong happend")
			continue
		}
		b, _ := agent.DefaultBsc.GetBlockByNumber(n)
		fmt.Printf("1. number:%d, miner:%s\n", n, b.Miner)
		if b.Difficulty != "0x1" {
			fmt.Println("RESTART LOOP: offturn slash should have difficulty=0x1, miner.Stop might fail")
			continue
		}
		agent.DefaultBsc.StartMiner()
		offTurnVal := b.Miner
		rTmp, ok := validators.CheckRecent(offTurnVal, idxCur)
		recs := NewRecents(len(validators) / 2)
		if ok {
			fmt.Println("validators:", *validators)
			fmt.Println("recents:", rTmp, "offTurnVal:", offTurnVal)
			panic("inturn validator slashed by recent validator")
		}
		for _, rv := range rTmp {
			recs.Push(rv)
		}
		recs.Push(b.Miner)
		nextBlockBySlash := validators.CalBlockNumberBySlash(n, offTurnVal, idxCur)
		CheckSlash(n, nextBlockBySlash, recs, validators, offTurnVal)
	}
}

func CheckSlash(nPrev, nextBlockBySlash int64, recs *recents, validators *vals, offTurnVal string) {
	nCum := 0
	for {
		n, _ := agent.DefaultBsc.GetBlockNumber()
		if n == nPrev {
			time.Sleep(1 * time.Second)
			continue
		}
		if n > nPrev+1 {
			fmt.Println("RESTART LOOP: something wrong happend during slash checking, skip blocks when scanning")
			return
		}
		nPrev = n
		b, _ := agent.DefaultBsc.GetBlockByNumber(n)
		fmt.Printf("2. number:%d, miner:%s\n", n, b.Miner)
		if b.Difficulty == "0x2" {
			recs.Push(b.Miner)
			nCum++
			if nCum >= len(validators) {
				fmt.Println("SLASH RECOVERD: restart loop")
				return
			}
		} else {
			nCum = 0
			if n < nextBlockBySlash {
				fmt.Println("RESTART LOOP: found some offTurn block happend unexpected, expected:", nextBlockBySlash, "found:", n)
				return
			}
			if n == nextBlockBySlash {
				miner := b.Miner
				if recs.CheckRecent(miner) {
					panic("recent signed happend when check about slash")
				}
				if recs.CheckRecent(offTurnVal) {
					//3s
					bPrev, _ := agent.DefaultBsc.GetBlockByNumber(n - 1)
					//t1, _ := strconv.Atoi(b.Timestamp[2:])
					//t2, _ := strconv.Atoi(bPrev.Timestamp[2:])
					t1, _ := strconv.ParseInt(b.Timestamp[2:], 16, 64)
					t2, _ := strconv.ParseInt(bPrev.Timestamp[2:], 16, 64)

					if t1-t2 == 3 {
						fmt.Println("SLASH: inturn validator has recnetSigned for earlier slash, this block be committed by offturn with 3s interval, blockNubmer", n)
					} else {
						fmt.Println(fmt.Sprintf("***********slash: should be 3s interval, but actually:%d on blockNumber:%d**********", t1-t2, n))
					}
				} else {
					//4s
					bPrev, _ := agent.DefaultBsc.GetBlockByNumber(n - 1)
					//					t1, _ := strconv.Atoi(b.Timestamp[2:])
					//					t2, _ := strconv.Atoi(bPrev.Timestamp[2:])
					t1, _ := strconv.ParseInt(b.Timestamp[2:], 16, 64)
					t2, _ := strconv.ParseInt(bPrev.Timestamp[2:], 16, 64)

					if t1-t2 == 4 {
						fmt.Println("SLASH: inturn validator has been slashed and not recentSigned, this block be committed by offturn with 4s interval, blockNubmer", n)
					} else {
						fmt.Println(fmt.Sprintf("**************slash: should be 4s interval, but actually:%d, on blockNumber:%d***************", t1-t2, n))
					}
				}
				idxTmp := 0
				for idxT, vT := range validators {
					if vT == offTurnVal {
						idxTmp = idxT
						break
					}
				}
				nextBlockBySlash = validators.CalBlockNumberBySlash(n, b.Miner, idxTmp)
				offTurnVal = miner
				recs.Push(miner)
			} else {
				fmt.Println("RESTART LOOP: unexpected slash")
			}
		}
	}
}

func WaitForVal(prevN int64, val string, validators *vals) (int64, *vals) {
	for {
		n, err := agent.DefaultBsc.GetBlockNumber()
		if err != nil || n == prevN {
			time.Sleep(1 * time.Second)
			continue
		}
		if n > prevN+1 {
			validators, prevN = GetValidators()
			preVal, _ := validators.GetPrev(valCurrent)
			return WaitForVal(prevN, preVal, validators)
		}
		prevN = n
		b, err := agent.DefaultBsc.GetBlockByNumber(n)
		fmt.Printf("3. number:%d, miner:%s\n", n, b.Miner)
		if err != nil {
			//	time.Sleep(1 * time.Second)
			continue
		}
		if b.Difficulty != "0x2" {
			validators, prevN = GetValidators()
			preVal, _ := validators.GetPrev(valCurrent)
			return WaitForVal(prevN, preVal, validators)
		}
		if b.Miner == val {
			return n, validators
		}
		time.Sleep(1 * time.Second)
	}
}

func GetValidators() (*vals, int64) {
	n, err := agent.DefaultBsc.GetBlockNumber()

	if err != nil {
		panic(err)
	}
	//	validators := [7]string{}
	validators := vals{}
	prevN := n
	index := 0
	b, err := agent.DefaultBsc.GetBlockByNumber(n)
	tPrev, _ := strconv.ParseInt(b.Timestamp[2:], 16, 64)
	for {
		n, err := agent.DefaultBsc.GetBlockNumber()
		if n == prevN {
			time.Sleep(1 * time.Second)
			continue
		}
		if n > prevN+1 {
			prevN = n
			index = 0
			time.Sleep(1 * time.Second)
			continue
		}
		prevN = n
		b, err = agent.DefaultBsc.GetBlockByNumber(n)
		t, _ := strconv.ParseInt(b.Timestamp[2:], 16, 64)
		fmt.Printf("4. number:%d, miner:%s, timestampDelta:%d\n", n, b.Miner, t-tPrev)
		tPrev = t
		if err != nil || b.Difficulty != "0x2" {
			fmt.Printf("GetValidators: err:%v, index:%d, difficulty:%s, blockNumber:%d\n", err, index, b.Difficulty, n)
			index = 0
			time.Sleep(1 * time.Second)
			continue
		}
		//		validators[index] = b.Miner
		validators.Update(index, b.Miner)
		index++
		if index == 5 {
			return &validators, n
		}
		time.Sleep(1 * time.Second)
	}
}
