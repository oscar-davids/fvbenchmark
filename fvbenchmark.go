package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"

	"github.com/golang/glog"
	"github.com/livepeer/lpms/ffmpeg"
)

func selectFile(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || (filepath.Ext(path) != ".bin" && filepath.Ext(path) != ".hash") {
			return nil
		}
		*files = append(*files, path)
		return nil
	}
}
func perf_measure(y_actual []int, y_infer []int) (int, int, int, int) {
	TP := 0
	FP := 0
	TN := 0
	FN := 0

	for i := 0; i < len(y_actual); i++ {
		if y_actual[i] == y_infer[i] && y_infer[i] == 1 {
			TP += 1
		}
		if y_infer[i] == 1 && y_actual[i] != y_infer[i] {
			FP += 1
		}
		if y_actual[i] == y_infer[i] && y_infer[i] == 0 {
			TN += 1
		}
		if y_infer[i] == 0 && y_actual[i] != y_infer[i] {
			FN += 1
		}
	}

	return TP, FP, TN, FN
}

func main() {
	glog.Info("hi")

	//indir1 := os.Args[0]
	//indir2 := os.Args[1]
	indir1 := "./hashdata/fvgpu/"
	indir2 := "./hashdata/fvcpu/"

	if indir1 == "" || indir2 == "" {
		panic("please select valid mpeg-7 hash folders")
	}
	var infiles1 []string
	err1 := filepath.Walk(indir1, selectFile(&infiles1))
	if len(infiles1) == 0 || err1 != nil {
		panic("Can not collect fail case files")
	}
	sort.Strings(infiles1)

	var infiles2 []string
	err2 := filepath.Walk(indir2, selectFile(&infiles2))
	if len(infiles2) == 0 || err2 != nil {
		panic("Can not collect fail case files")
	}
	sort.Strings(infiles2)

	fmt.Println("Task starting.")

	// create csv file
	outcsv := "compresult.csv"
	fwriter, err := os.Create(outcsv)
	defer fwriter.Close()
	if err != nil {
		panic("Can not create csv file")
	}
	csvrecorder := csv.NewWriter(fwriter)
	defer csvrecorder.Flush()
	//write header
	columnheader := []string{"filepath1", "filepath2", "real", "infer"}
	_ = csvrecorder.Write(columnheader)

	// true comparison
	paircount := len(infiles1)
	var halfpair int
	halfpair = paircount / 2
	truelabel := []int{}
	predict := []int{}

	for i := 0; i < paircount; i++ {
		bequal, _ := ffmpeg.CompareSignatureByPath(infiles1[i], infiles2[i])

		truelabel = append(truelabel, 1)
		sfequal := "1"
		if !bequal {
			sfequal = "0"
			predict = append(predict, 0)
		} else {
			predict = append(predict, 1)
		}

		var linestr []string
		_, filename1 := filepath.Split(infiles1[i])
		_, filename2 := filepath.Split(infiles2[i])
		linestr = append(linestr, filename1)
		linestr = append(linestr, filename2)
		linestr = append(linestr, "1")
		linestr = append(linestr, sfequal)
		csvrecorder.Write(linestr)
		fmt.Println("current ", i, filename1, filename2)
	}

	// false comparison
	for i := 0; i < paircount; i++ {
		j := 0
		if i < halfpair {
			j = halfpair + rand.Intn(halfpair)
		} else {
			j = rand.Intn(halfpair)
		}
		bequal, _ := ffmpeg.CompareSignatureByPath(infiles1[i], infiles2[j])

		truelabel = append(truelabel, 0)
		sfequal := "0"
		if bequal {
			sfequal = "1"
			predict = append(predict, 1)
		} else {
			predict = append(predict, 0)
		}
		var linestr []string
		_, filename1 := filepath.Split(infiles1[i])
		_, filename2 := filepath.Split(infiles2[j])
		linestr = append(linestr, filename1)
		linestr = append(linestr, filename2)
		linestr = append(linestr, "0")
		linestr = append(linestr, sfequal)
		csvrecorder.Write(linestr)
		fmt.Println("current ", i, filename1, filename2)
	}
	//calculate accuracy and false positive & false negative
	TP, FP, TN, FN := perf_measure(truelabel, predict)

	ACC := float64(TP+TN) / float64(TP+FP+FN+TN)
	// Fall out or false positive rate
	FPR := float64(FP) / float64(FP+TN)
	// False negative rate
	FNR := float64(FN) / float64(TP+FN)
	// Sensitivity, hit rate, recall, or true positive rate
	TPR := float64(TP) / float64(TP+FN)
	//Specificity or true negative rate
	TNR := float64(TN) / float64(TN+FP)

	fmt.Printf("=========================================\n")
	fmt.Printf("=========================================\n")
	fmt.Printf("true positive := %v, true negative := %v\n", TPR, TNR)
	fmt.Printf("false positive := %v, False negative := %v\n", FPR, FNR)
	fmt.Printf("=========================================\n")
	fmt.Printf("Accuracy := %v\n", ACC)
	fmt.Printf("=========================================\n")

	fmt.Printf("Task completed!")

}
