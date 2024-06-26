package bloomfilter

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"hash"
	"log"
	"math"
	"os"

	"github.com/spaolacci/murmur3"
)

// The first four bytes will be an identifier, we’ll use CCBF.(Id)
// The next two bytes will be a version number to describe the version number of the file. (V)
// The next two bytes will be the number of hash functions used. (K)
// The next four bytes will be the number of bits used for the filter. (M)

type Bloom struct {
	M      int32  // size of bitset
	K      uint32 // no.of hash functions
	Bitset []bool
	Hashes []hash.Hash32
}

// n is size of bitset
func (bloomFilter *Bloom) New(n int, p float64) {
	bloomFilter.M = M(n, p)
	bloomFilter.K = K(bloomFilter.M, n)
	bloomFilter.Bitset = make([]bool, bloomFilter.M)
	bloomFilter.Hashes = make([]hash.Hash32, bloomFilter.K)
	var i uint32
	for i = 0; i < bloomFilter.K; i++ {
		bloomFilter.Hashes[i] = murmur3.New32WithSeed(i)
	}
}

func (bloomFilter *Bloom) Add(word string) {
	for _, hashFunc := range bloomFilter.Hashes {
		hashFunc.Reset()
		hashFunc.Write([]byte(word))
		idx := hashFunc.Sum32() % uint32(bloomFilter.M)
		bloomFilter.Bitset[idx] = true
	}
}

func (bloomFilter *Bloom) Contains(word string) bool {
	for _, hashFunc := range bloomFilter.Hashes {
		hashFunc.Reset()
		hashFunc.Write([]byte(word))
		idx := hashFunc.Sum32() % uint32(bloomFilter.M)
		if !bloomFilter.Bitset[idx] {
			return false
		}
	}
	return true
}

/*
M = -((n * ln(p)) / (ln(2) * ln(2)))
K = m / n * ln(2)
*/
func M(n int, p float64) int32 {
	return int32(math.Ceil(-1 * float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
}

func K(m int32, n int) uint32 {
	return uint32(math.Ceil(float64(m) * math.Log(2) / float64(n)))
}

func BuildBloomFilter(inputFile string, outputFilePath string, p float64) error {
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("err while opening dict file: %v", err)
	}

	defer file.Close()

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Fatalf("err while opening dict file: %v", err)
	}

	defer outputFile.Close()

	scanner := bufio.NewScanner(file)

	words := []string{}
	for scanner.Scan() {
		word := scanner.Text()
		words = append(words, word)
	}

	n := len(words)
	fmt.Println("no.of words:", n)

	b := &Bloom{}
	b.New(n, p)
	for _, word := range words {
		b.Add(word)
	}

	identifier := []byte("CCBF")
	_, err = outputFile.Write(identifier)
	if err != nil {
		return err
	}

	version := uint16(1)
	err = binary.Write(outputFile, binary.LittleEndian, version)
	if err != nil {
		return err
	}

	err = binary.Write(outputFile, binary.LittleEndian, b.K)
	if err != nil {
		return err
	}

	err = binary.Write(outputFile, binary.LittleEndian, b.M)
	if err != nil {
		return err
	}

	for _, bit := range b.Bitset {
		var bitValue byte
		if bit {
			bitValue = 1
		} else {
			bitValue = 0
		}

		err = binary.Write(outputFile, binary.LittleEndian, bitValue)
		if err != nil {
			return err
		}
	}

	return nil
}

func LoadBloomFilter(filePath string) (*Bloom, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	b := &Bloom{}
	identifier := make([]byte, 4)
	var version uint16

	_, err = file.Read(identifier)
	if err != nil {
		return nil, err
	}

	err = binary.Read(file, binary.LittleEndian, &version)
	if err != nil {
		return nil, err
	}

	err = binary.Read(file, binary.LittleEndian, &b.K)
	if err != nil {
		return nil, err
	}

	err = binary.Read(file, binary.LittleEndian, &b.M)
	if err != nil {
		return nil, err
	}

	b.Bitset = make([]bool, b.M)
	for i := 0; i < int(b.M); i++ {
		var byteValue byte
		err = binary.Read(file, binary.LittleEndian, &byteValue)
		if err != nil {
			return nil, err
		}
		if byteValue == 1 {
			b.Bitset[i] = true
		} else {
			b.Bitset[i] = false
		}
	}

	b.Hashes = make([]hash.Hash32, b.K)
	for i := 0; i < int(b.K); i++ {
		b.Hashes[i] = murmur3.New32WithSeed(uint32(i))
	}

	return b, nil
}

func SpellCheck(filename string, args []string) {
	b, err := LoadBloomFilter(filename)
	if err != nil {
		fmt.Println(err)
	}

	output := "These words are spelt wrong:"
	for _, word := range args {
		// fmt.Printf("%s -> %v", word, b.Contains(word))
		if !b.Contains(word) {
			output += fmt.Sprintf("\n %s", word)
		}
	}

	fmt.Println(output)
}
