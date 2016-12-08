package util

import (
	"crypto/rand"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	ds "github.com/enirinth/blob-storage/clusterds"
	"io"
	"strconv"
	"os"
	"io/ioutil"
	"strings"
	mrand "math/rand"
	"text/tabwriter"
)

const (
	separator string  = "-----"
	EPSILON   float64 = 0.00000001
)

func FloatEquals(a, b float64) bool {
	if (a-b) < EPSILON && (b-a) < EPSILON {
		return true
	}
	return false
}

// UUid generator
// newUUID generates a random UUID according to RFC 4122
// Credit to https://play.golang.org/p/4FkNSiUDMg
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// Print storage table for a certain DC
func PrintStorage(storageTable *map[string]*ds.Partition) {
	fmt.Println("#### Storage Map ###")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Fprintln(w, "PartitionID\tBlob List\tPartition Size\tTimestamp")

	for _, v := range *storageTable {
		line := (*v).PartitionID + "\t"
		tmp := ""
		for _, blob := range (*v).BlobList {
			tmp += "{" + blob.BlobID + ", " + strconv.FormatFloat(blob.BlobSize, 'f', 6, 64) + "}, "
		}
		line += tmp + "\t"
		line += strconv.FormatFloat((*v).PartitionSize, 'f', 6, 64) + "\t"
		line += strconv.FormatInt((*v).CreateTimestamp, 10)
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

func PrintCluster(ReplicaMap *map[string]*ds.PartitionState, ReadMap *map[string]*ds.NumRead){
	fmt.Println("### Cluster Map ###")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Fprintln(w, "PartitionID\tDC List\tlocal Read\tGlobal Read\t")

	for partitionID, DCs := range *ReplicaMap {
		line := partitionID + "\t"
		tmp := ""
		for _, name := range DCs.DCList {
			tmp += name + ","
		}
		line += tmp + "\t"
		lRead := (*ReadMap)[partitionID].LocalRead
		gRead := (*ReadMap)[partitionID].GlobalRead
		line += strconv.Itoa(int(lRead)) + "\t" + strconv.Itoa(int(gRead)) + "\t"
		fmt.Fprintln(w, line)
	}
	w.Flush()
}

func ReadFile(filename string) []string {
	fmt.Println(os.Getwd())
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(dat), "\n")
	return lines
}


// pick a random number from [0, numDC); return the dc name as string
func GetRandomDC(numDC int) string{
	random := mrand.Intn(numDC)
	return strconv.Itoa(random + 1)
}


// pick a random DC from the DCList; return the dc name as string
func GetRandomDCFromList(dcList []string) string {
	random := mrand.Intn(len(dcList))
	return dcList[random]
}


// Find if a certain DC stores a certain partition
func FindDC(dcID string, pState *ds.PartitionState) bool {
	for _, id := range (*pState).DCList {
		if id == dcID {
			return true
		}
	}
	return false
}

// Find blob (using its ID) in a partition
// return True if found
func FindBlob(blobID string, partition *ds.Partition) bool {
	for _, blob := range partition.BlobList {
		if blobID == blob.BlobID {
			return true
		}
	}
	return false
}

// Merge one partition(p2) into another(p1), so that it(p1) contains (in set semantics) all the blobs
// Happens during inter-DC synchronization
func MergePartition(p1 *ds.Partition, p2 *ds.Partition) {
	if p1.PartitionID != p2.PartitionID {
		err := errors.New("Cannot merge two different partitions (with different IDs)")
		log.Fatal(err)
	}
	for _, blob := range p2.BlobList {
		if !FindBlob(blob.BlobID, p1) {
			p1.AppendBlob(blob)
			p1.PartitionSize += blob.BlobSize
		}
	}
}

func FindPartition(partitionID string, m *map[string]*ds.PartitionState) bool {
	_, ok := (*m)[partitionID]
	return ok
}
