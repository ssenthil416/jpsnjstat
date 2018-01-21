package jpsnjstat

import (
	"bytes"
	"errors"
	_ "fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	_ "time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
	javaHome = "/usr/local/java"
`

type Metric struct {
	ProcessID string  `json:"processID"`
	S0C       float32 `json:"s0c"`
	S1C       float32 `json:"s1c"`
	S0U       float32 `json:"s0u"`
	S1U       float32 `json:"s1u"`
	EC        float32 `json:"ec"`
	EU        float32 `json:"eu"`
	OC        float32 `json:"oc"`
	OU        float32 `json:"ou"`
	MC        float32 `json:"mc"`
	MU        float32 `json:"mu"`
	CCSC      float32 `json:"ccsc"`
	CCSU      float32 `json:"ccsu"`
	YGC       float32 `json:"ygc"`
	YGCT      float32 `json:"ygct"`
	FGC       float32 `json:"fgc"`
	FGCT      float32 `json:"fgct"`
	GCT       float32 `json:"gct"`
}

type Jpsnjstat struct {
	JavaHome string `toml:"javaHome"`
	Metrics  []Metric
}

func (j *Jpsnjstat) SampleConfig() string {
	return sampleConfig
}

func (j *Jpsnjstat) Description() string {
	return "Read JVM metrics through jpsnjstat"
}

func getFPSData(path string) string {
	jpsCmd := "jps"
	jpsCmd = path + "/" + jpsCmd
	cmd := exec.Command(jpsCmd)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fpsOut := out.String()
	fpsOut = strings.TrimRight(fpsOut, "\n")
	//fmt.Println(" fps = ", fpsOut)
	return fpsOut
}

func getFstatData(arg2 string, path string) string {
	jstatComm := "jstat"
	arg1 := "-gc"
	arg3 := "1"
	jstatComm = path + "/" + jstatComm
	cmd := exec.Command(jstatComm, arg1, arg2, arg3, arg3)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	fstatOut := out.String()
	fstatOut = strings.TrimRight(fstatOut, "\n")
	//fmt.Println(" fstat = ", fstatOut)
	return fstatOut
}

func delete_empty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" && str != "\n" && str != "\r" {
			//fmt.Println("str =", str)
			r = append(r, str)
		}
	}
	return r
}

func insertToInfxDB(pID string, input string) map[string]interface{} {
	//fmt.Println("input =", input)
	fields := make(map[string]interface{})

	spInput := strings.Split(input, "\n")
	colName := strings.Split(spInput[0], " ")
	colName = delete_empty(colName)
	rownValue := strings.Split(spInput[1], " ")
	rownValue = delete_empty(rownValue)

	//fmt.Println("len =", len(colName), len(rownValue))

	fields["ProcessID"] = pID
	for k, _ := range colName {
		//fmt.Println(colName[k], " = ", rownValue[k])
		fields[colName[k]], _ = strconv.ParseFloat(rownValue[k], 32)
	}

	return fields
}

func (j *Jpsnjstat) Gather(acc telegraf.Accumulator) error {
	//startTime := time.Now()
	//fmt.Println("Time = ", startTime)
	fpsStr := getFPSData(j.JavaHome)
	lineFpsStr := strings.Split(fpsStr, "\n")
	//fmt.Println("lineFpsStr =", lineFpsStr)
	//fmt.Println("len of lineFpsStr =", len(lineFpsStr))
	listFpsPs := make([]string, 0, 1)
	for _, v := range lineFpsStr {
		//fmt.Println("v =", v)
		if strings.Contains(v, "Jps") == false {
			spToPsID := strings.Split(v, " ")
			listFpsPs = append(listFpsPs, spToPsID[0])
		}
	}

	if len(listFpsPs) < 0 {
		log.Fatal(errors.New("No JVM process are running"))
		//break
	}

	for _, v := range listFpsPs {
		fstatOut := getFstatData(v, j.JavaHome)
		data := insertToInfxDB(v, fstatOut)
		acc.AddFields("jpsnjstat", data, nil)
	}
	return nil
}

func init() {
	inputs.Add("jpsnjstat", func() telegraf.Input {
		return &Jpsnjstat{}
	})
}
