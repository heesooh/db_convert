package main

import (
	"os"
	"fmt"
	"time"
	"flag"
	"errors"
	"strconv"
	"github.com/dappley/go-dappley/util"
	"github.com/dappley/go-dappley/storage"
	"github.com/dappley/go-dappley/core/block"
)

var tipKey = []byte("tailBlockHash")

func main() {
	//checks if the flag is valid
	args := os.Args[1:]
	if len(args) < 1 {
		printUsage()
		return
	}

	var fileName string
	flag.StringVar(&fileName, "file", "default.db", "default database file")
	flag.Parse()

	//load file and convert
	convertDBtoSQL(fileName)
}

//prints out the usage
func printUsage() {
	fmt.Println("----------------------------------------------")
	fmt.Println("Usage: Converts the .db file into a .sql file.")
	fmt.Println("Usage Example: ./db_convert -file default.db")
}

//checks whether the file exists or not
func isDbExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

//loads db file info
func LoadDBFile(fileName string) (storage.Storage, error) {
	isExist := isDbExist(fileName)
	if !isExist {
		err := errors.New("File " + fileName + " does not exist")
		return nil, err
	}
	db := storage.OpenDatabase(fileName)
	return db, nil
}

//return the height of the last block in the blockchain
func lastBlockHeight(db storage.Storage) uint64 {
	hash, err := db.Get(tipKey)
	if err != nil {
		panic(err)
	}
	rawBytes, err := db.Get(hash)
	if err != nil {
		panic(err) 
	}
	return block.Deserialize(rawBytes).GetHeight()
}

func orderedData(block *block.Block) string {
	id    := strconv.FormatUint(block.GetHeight(), 10)
	index := id
	hash  := block.GetHash().String()
	pre_hash  := block.GetPrevHash().String()
	nonce := "0"
	timestamp := block.GetTimestamp()
	transactions := "transactions"
	miner := block.GetProducer()
	size  := "size"
	crt_time := time.Now().Format("2006-01-02 15:04:05") //time.Unix(timestamp, 0).String() 

	ordered_data := ("(" + id + ", " + index + ", '" + hash + "', '" + pre_hash + 
					 "', " + nonce + ", " + strconv.FormatInt(timestamp, 10) + ", '" + transactions + 
					 "', '" + miner + "', " + size + ", '" + crt_time + "'),")
	return ordered_data
}

//iterate through the blockchain to print the inromation in order
func convert(db storage.Storage, tailHeight uint64) string {
	var data string
	insert_info := "INSERT INTO `t_block` (`id`, `index`, `hash`, `pre_hash`, " +
				  "`nonce`, `timestamp`, `transactions`, `miner`, `size`, `crt_time`) VALUES\n"
	for i := uint64(1); i <= tailHeight; i++ {
		hash, err := db.Get(util.UintToHex(i))
		if err != nil {
			fmt.Println(err)
			return "Block at height " + strconv.FormatUint(i, 10) + " does not exist"
		}
		rawBytes, err := db.Get(hash)
		block := block.Deserialize(rawBytes)
		ordered_data := orderedData(block) + "\n"
		if i % 164 == 1 {
			data += insert_info + ordered_data
		} else {
			data += ordered_data
		}
	}
	return data
}

//convert the .db file to .sql file
func convertDBtoSQL(fileName string) {
	db, err := LoadDBFile(fileName)
	if err != nil {
        panic(err)
    }
    defer db.Close()
	
	tailHeight := lastBlockHeight(db)
	data := convert(db, tailHeight)

	//fmt.Println(data)

	// //create new file and write data in the file
	file, err := os.Create("dappleyweb.sql")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.Write([]byte(data))
	if err != nil {
		panic(err)
	}
}