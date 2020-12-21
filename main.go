package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"io"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
)

type Suspect struct {
	Time    *big.Int
	Address string
	Amount  *big.Int
}

type IncomingInfo struct {
	TimeSent   *big.Int
	AmountSent *big.Int
}

func main() {
	maximum, minimum := calculateMaximum()
	topDivider := generateTopDivider(maximum, minimum)
	var step big.Int
	step.Div(&topDivider, big.NewInt(10))
	generateHystorgram(topDivider, step, "out.csv")
	info := make([]IncomingInfo, 4)
	var comissionedAmount big.Int
	comissionedAmount.Mul(big.NewInt(164744645000000000), big.NewInt(int64(0.97*10000)))
	comissionedAmount.Div(&comissionedAmount, big.NewInt(10000))
	fmt.Println("Comissioned Amount: ", comissionedAmount.String())
	info[0] = IncomingInfo{
		TimeSent:   big.NewInt(1608199138),
		AmountSent: big.NewInt(164744645000000000),
	}
	info[1] = IncomingInfo{
		TimeSent:   big.NewInt(1608076800),
		AmountSent: big.NewInt(164744645000000000),
	}
	info[2] = IncomingInfo{
		TimeSent:   big.NewInt(1607990400),
		AmountSent: big.NewInt(164744645000000000),
	}
	info[3] = IncomingInfo{
		TimeSent:   big.NewInt(1607986800),
		AmountSent: big.NewInt(164744645000000000),
	}

	res := findSuspects(info, 0.05, 0.01, big.NewInt(24*60*60))
	for k, v := range res {
		if !v {
			continue
		}
		fmt.Println(k)
	}
	fmt.Println(len(res))

}
func findSuspects(info []IncomingInfo, comissionWindowWorst, comissionWindowBest float64, timeWindow *big.Int) map[string]bool {
	suspectsByInfo := make([][]Suspect, len(info))
	inputData, err := os.Open("datatiming.csv")
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(inputData)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		splitted := strings.Split(string(line), ",")
		var time, amount big.Int
		name := splitted[0]
		getFromString(&amount, splitted[1])
		getFromString(&time, splitted[2])
		for i := 0; i < len(info); i++ {
			var lowerAmountGap, higherAmountGap, lowerTimeGap, higherTimeGap big.Int
			lowerAmountGap.Mul(info[i].AmountSent, big.NewInt(int64((1-comissionWindowWorst)*10000)))
			lowerAmountGap.Div(&lowerAmountGap, big.NewInt(10000))

			higherAmountGap.Mul(info[i].AmountSent, big.NewInt(int64((1-comissionWindowBest)*10000)))
			higherAmountGap.Div(&higherAmountGap, big.NewInt(10000))

			lowerTimeGap.Set(info[i].TimeSent)
			higherTimeGap.Add(info[i].TimeSent, timeWindow)
			//fmt.Println(i, " ", lowerAmountGap.String(), " ", higherAmountGap.String(), " ", lowerTimeGap.String(), " ", higherTimeGap.String())
			if amount.Cmp(&lowerAmountGap) == 1 && amount.Cmp(&higherAmountGap) == -1 &&
				time.Cmp(&lowerTimeGap) == 1 && time.Cmp(&higherTimeGap) == -1 {
				suspectsByInfo[i] = append(suspectsByInfo[i], Suspect{
					Time:    &time,
					Amount:  &amount,
					Address: name,
				})
			}
		}
	}
	sets := make([]map[string]bool, len(suspectsByInfo))
	for i := 0; i < len(suspectsByInfo); i++ {
		sets[i] = make(map[string]bool)
		for j := 0; j < len(suspectsByInfo[i]); j++ {
			sets[i][suspectsByInfo[i][j].Address] = true
		}
	}
	fmt.Println("LENGHTS:")
	for i := 0; i < len(suspectsByInfo); i++ {
		fmt.Println("Amount of suspects in ", i, ": ", len(sets[i]))
		fmt.Println("With duplicates in ", i, " :", len(suspectsByInfo[i]))
	}
	res := make(map[string]bool)
	for i := 0; i < len(sets); i++ {
		for key, value := range sets[i] {
			isThere := true
			if value == false {
				continue
			}
			for k := 0; k < len(sets); k++ {
				if k == i  {
					continue
				}
				isThere = isThere && sets[k][key]
			}
			if isThere {
				res[key] = true
			}
		}
	}

	err = inputData.Close()
	if err != nil {
		panic(err)
	}
	return res
}

func contains(suspects []Suspect, suspect Suspect) bool {
	for i:=0; i < len(suspects); i++ {
		if suspects[i].Address == suspect.Address {
			return true
		}
	}
	return false
}

func generateHystorgram(topDivider, step big.Int, fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	inputData, err := os.Open("datatiming.csv")
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(inputData)
	//writer := bufio.NewWriter(file)
	results := make(map[int64]int64)
	maxIndex := int64(0)
	var value big.Int
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		splitted := strings.Split(string(line), ",")
		getFromString(&value, splitted[1])
		if value.Cmp(&topDivider) == -1 {
			var res big.Int
			res.Div(&value, &step)
			results[res.Int64()]++
			if res.Int64() > maxIndex {
				maxIndex = res.Int64()
			}
		}
	}
	fmt.Println("Max index: ", maxIndex)
	for i := int64(0); i <= maxIndex; i++ {
		fmt.Println(i, " ", results[i])
	}


	err = file.Close()
	if err != nil {
		panic(err)
	}
	err = inputData.Close()
	if err != nil {
		panic(err)
	}
}

func getFromString(number *big.Int, input string) {
	_, success := number.SetString(input, 10)
	if !success {
		panic(errors.New("problem getting number from string"))
	}
}

func generateTopDivider(maximum, minimum big.Int) big.Int {
	inputData, err := os.Open("datatiming.csv")
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(inputData)
	var value, dividedMaximum big.Int
	counter := 0
	otherCounter := 0
	divider := int64(10000000)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		splitted := strings.Split(string(line), ",")
		getFromString(&value, splitted[1])
		if value.Cmp(dividedMaximum.Div(&maximum, big.NewInt(divider))) == -1 {
			counter++
		} else {
			otherCounter++
		}
	}
	fmt.Println("Divided maximum: ", dividedMaximum.String())
	fmt.Println("transactions less than max /", divider, ": ", counter)
	fmt.Println("transactions bigger than max /", divider, ": ", otherCounter)
	fmt.Printf("trimmed percent: %f percent \n", float32(otherCounter*100)/float32(counter+otherCounter))
	err = inputData.Close()
	if err != nil {
		panic(err)
	}
	return dividedMaximum
}

func calculateMaximum() (big.Int, big.Int) {
	inputData, err := os.Open("datatiming.csv")
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(inputData)
	line, _, err := reader.ReadLine()
	if err != nil {
		panic(err)
	}
	splitted := strings.Split(string(line), ",")
	var maximum, minimum, value big.Int
	getFromString(&maximum, splitted[1])
	getFromString(&minimum, splitted[1])
	fmt.Println(maximum.String())
	fmt.Println(minimum.String())
	counter := 0
	for {
		line, _, err := reader.ReadLine()
		counter++
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		splitted := strings.Split(string(line), ",")
		getFromString(&value, splitted[1])
		if value.Cmp(&maximum) == 1 {
			maximum.Set(&value)
		}
		if value.Cmp(&minimum) == -1 {
			minimum.Set(&value)
		}
	}
	//totalRecieved := make(map[string]int)
	fmt.Println("amount of transactions: ", counter)
	fmt.Println("maximum: ", maximum.String())
	fmt.Println("minimum: ", minimum.String())
	err = inputData.Close()
	if err != nil {
		panic(err)
	}
	return maximum, minimum
}

func getData() {
	fmt.Println("Hello world")
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/ea0a6fb4db1d4e2fbe565b60d324d2fa")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("we have a connection")
	f, err := os.Create("datatiming.csv")
	if err != nil {
		panic(err)
	}
	num := 11491004
	globalCount := 0
	for i := num; i != num-40000; i-- {
		block, err := client.BlockByNumber(context.Background(), big.NewInt(int64(i)))
		fmt.Println(i)
		if err != nil {
			fmt.Println(err)
			continue
		}
		transactions := block.Transactions()
		blockResult := ""
		count := 0
		for j := 0; j < transactions.Len(); j++ {
			to := transactions[j].To()
			value := transactions[j].Value()
			if value.Int64() == 0 || to == nil || to.String() == "" || value == nil || value.String() == "" {
				continue
			}
			blockResult += to.String() + "," + value.String() + "," + strconv.FormatUint(block.Time(), 10) + "\n"
			count++
		}
		fmt.Println("Block ", i, " not nil transactions ", count, " sum ", globalCount)
		_, err = f.WriteString(blockResult)
		if err != nil {
			fmt.Println(err)
		}
		globalCount++
	}
}
