package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// BloodUnit represents a unit of blood in the blockchain
type BloodUnit struct {
	UnitID       string `json:"unit_id"`
	DonorID      string `json:"donor_id"`
	BloodType    string `json:"blood_type"`
	DonationDate string `json:"donation_date"`
	ExpiryDate   string `json:"expiry_date"`
	Status       string `json:"status"` // Donated, Stored, Used, Expired
	RecipientID  string `json:"recipient_id"`
	IsGenesis    bool   `json:"isgenesis"`
}

// Blocks represents a block in the blockchain
type Blocks struct {
	Pos       int       `json:"pos"`
	Data      BloodUnit `json:"data"`
	Hash      string    `json:"hash"`
	TimeStamp string    `json:"time"`
	PrevHash  string    `json:"prevhash"`
}

// Blockchain is a series of validated Blocks
type Blockchain struct {
	blocks []*Blocks
}

// Global variable to hold the blockchain
var BlockChain *Blockchain

// AddBlock adds a new block to the blockchain
func (bc *Blockchain) AddBlock(info BloodUnit) {
	prevBlock := bc.blocks[len(bc.blocks)-1]

	block := CreateBlock(prevBlock, info)

	if validBlock(block, prevBlock) {
		bc.blocks = append(bc.blocks, block)
	} else {
		log.Println("Failed to append block")
	}
}

// GenesisBlock creates the first block in the blockchain
func GenesisBlock() *Blocks {
	return CreateBlock(&Blocks{}, BloodUnit{IsGenesis: true})
}

// NewBlockchain creates a new Blockchain with initialized blocks
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Blocks{GenesisBlock()}}
}

// CreateBlock creates a new block using previous block and transaction data
func CreateBlock(prevBlock *Blocks, info BloodUnit) *Blocks {
	block := &Blocks{}
	block.Pos = prevBlock.Pos + 1
	block.Data = info
	block.PrevHash = prevBlock.Hash
	block.TimeStamp = time.Now().String()
	block.GenerateHash()
	return block
}

// GenerateHash generates a hash for the block
func (b *Blocks) GenerateHash() {
	bytes, _ := json.Marshal(b.Data)
	data := fmt.Sprintf("%d%s%s%s", b.Pos, b.TimeStamp, string(bytes), b.PrevHash)
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

// validBlock validates the block with respect to the previous block
func validBlock(block, prevBlock *Blocks) bool {

	if prevBlock.Hash != block.PrevHash {
		return false
	}

	if !block.validateHash(block.Hash) {
		return false
	}

	if prevBlock.Pos+1 != block.Pos {
		return false
	}
	return true
}

// validateHash checks whether the hash of the block is correct
func (b *Blocks) validateHash(hash string) bool {
	b.GenerateHash()
	if b.Hash != hash {
		return false
	}
	return true
}

// CreateBloodEntry adds a new blood unit to the blockchain
func CreateBloodEntry(c *gin.Context) {
	var bloodUnit BloodUnit

	err := c.BindJSON(&bloodUnit)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}

	h := md5.New()
	uniqueString := bloodUnit.DonorID + bloodUnit.BloodType + bloodUnit.DonationDate
	io.WriteString(h, uniqueString)
	bloodUnit.UnitID = fmt.Sprintf("%x", h.Sum(nil))

	BlockChain.AddBlock(bloodUnit)

	c.IndentedJSON(http.StatusCreated, bloodUnit)

}

// GetBlockchain returns the current state of the blockchain
func GetBlockchain(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, BlockChain.blocks)
}

func main() {
	BlockChain = NewBlockchain()
	r := gin.Default()

	r.POST("/bloodentry", CreateBloodEntry)
	r.GET("/blockchain", GetBlockchain)

	// Debugging routine to print the state of the blockchain
	go func() {
		for _, block := range BlockChain.blocks {
			fmt.Printf("Prev. hash: %x\n", block.PrevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", " ")
			fmt.Printf("Data: %v\n", string(bytes))
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Println()
		}
	}()

	fmt.Println("starting server at port 8080")
	r.Run("localhost:8080")
}
