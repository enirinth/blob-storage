package util

import (
	"crypto/rand"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	config "github.com/enirinth/blob-storage/clusterconfig"
	ds "github.com/enirinth/blob-storage/clusterds"
	"io"
	"io/ioutil"
	"math"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
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

func PrintCluster(ReplicaMap *map[string]*ds.PartitionState, ReadMap *map[string]*ds.NumRead) {
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
	//fmt.Println(os.Getwd())
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(dat), "\n")
	return lines
}

// pick a random number from [0, numDC); return the dc name as string
func GetRandomDC(numDC int) string {
	random := mrand.Intn(numDC)
	return strconv.Itoa(random + 1)
}

// pick a random DC from the DCList; return the dc name as string
func GetRandomDCFromList(dcList []string) string {
	random := mrand.Intn(len(dcList))
	return dcList[random]
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

// Returns the size of (total) difference of two partitions
// e.g. p1:blob1 blob2, p2: blob1 blob3; will return the size of blob2+blob3
func DeltaSize(p1 *ds.Partition, p2 *ds.Partition) float64 {
	if p1.PartitionID != p2.PartitionID {
		err := errors.New("Cannot calculate delta size of two different " +
			"partitions (with different IDs)")
		log.Fatal(err)
	}
	var deltaSize float64 = 0
	for _, blob := range p1.BlobList {
		if !FindBlob(blob.BlobID, p2) {
			deltaSize += blob.BlobSize
		}
	}
	for _, blob := range p2.BlobList {
		if !FindBlob(blob.BlobID, p1) {
			deltaSize += blob.BlobSize
		}
	}
	return deltaSize
}

// Find partition in a ReplicaMap
func FindPartition(partitionID string, m *map[string]*ds.PartitionState) bool {
	_, ok := (*m)[partitionID]
	return ok
}

// Simulate end-to-end latency of transfering a file
func MockTransLatency(dcID1 string, dcID2 string, size float64) {
	if !config.MockTransLatencyON {
		return
	}
	x := latencyOfSize(dcID1, dcID2, size)

	time.Sleep(time.Millisecond * time.Duration(x))
}

// Given 2 Datacenters and size of data in bytes to transfer, it returns the
// simulated latency
func latencyOfSize(dcID1 string, dcID2 string, size float64) int {
	ireland := config.DC1
	northVa := config.DC2
	northCal := config.DC3

	if dcID1 == northCal {
		if dcID2 == northVa {
			return latencyCalVirg(size)
		} else if dcID2 == ireland {
			return latencyIreCal(size)
		}
	} else if dcID1 == northVa {
		if dcID2 == northCal {
			return latencyCalVirg(size)
		} else if dcID2 == ireland {
			return latencyIreCal(size)
		}
	} else if dcID1 == ireland {
		if dcID2 == northCal {
			return latencyIreCal(size)
		} else if dcID2 == northVa {
			return latencyIreVirg(size)
		}
	} else {
		fmt.Print("latencyOfSize has wrong paramters", dcID1, dcID2, "\n")
		return 0
	}
	return 0
}

// Formula for N. Cal and N. Virg latency time
func latencyCalVirg(s float64) int {
	connectionTime := 150
	if s <= 2097152 {
		x := int(105.92*math.Log2(s)/math.Log2(math.E) - 840.64)
		if x < 150 {
			return 0
		} else {
			return (x - connectionTime)
		}
	} else {
		return (int(0.0001*s+515.34) - connectionTime)
	}

}

// Formula for Ireland and N. Virg latency time
func latencyIreVirg(s float64) int {
	connectionTime := 146
	if s <= 2097152 {
		x := int(113.15*math.Log2(s)/math.Log2(math.E) - 944)
		if x < 146 {
			return 0
		} else {
			return (x - connectionTime)
		}
	} else {
		return (int(.0001*s+502.88) - connectionTime)
	}
}

// Formula for Ireland and N. Cal latency time
func latencyIreCal(s float64) int {
	connectionTime := 292
	if s <= 2097152 {
		x := int(194.93*math.Log2(s)/math.Log2(math.E) - 1522)
		if x < 292 {
			return 0
		} else {
			return (x - connectionTime)
		}
	} else {
		return (int(.0002*s+869.06) - connectionTime)
	}
}
