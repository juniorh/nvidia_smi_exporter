package main

import (
    "bytes"
    "encoding/csv"
    "fmt"
    "net/http"
    "log"
    "os"
    "os/exec"
    "strings"
    "strconv"
)

func metrics(response http.ResponseWriter, request *http.Request) {
    out, err := exec.Command(
        "nvidia-smi",
        "--query-gpu=name,index,serial,count,temperature.gpu,utilization.gpu,utilization.memory,memory.total,memory.free,memory.used,encoder.stats.sessionCount,encoder.stats.averageFps,encoder.stats.averageLatency,power.draw,power.limit,power.min_limit,clocks.current.graphics,clocks.current.sm,clocks.current.memory,clocks.current.video,clocks.max.graphics,clocks.max.sm,clocks.max.memory,clocks.applications.graphics,clocks.applications.mem",
        "--format=csv,noheader,nounits").Output()

    if err != nil {
        fmt.Printf("%s\n", err)
        return
    }

    csvReader := csv.NewReader(bytes.NewReader(out))
    csvReader.TrimLeadingSpace = true
    records, err := csvReader.ReadAll()
    if err != nil {
        fmt.Printf("%s\n", err)
        return
    }

    metricList := []string {
        "count",
        "temperature.gpu", "utilization.gpu",
        "utilization.memory", "memory.total", "memory.free", "memory.used",
        "encoder.stats.sessionCount", "encoder.stats.averageFps", "encoder.stats.averageLatency", 
        "power.draw", "power.limit", "power.min_limit", 
        "clocks.current.graphics", "clocks.current.sm", "clocks.current.memory", "clocks.current.video", 
        "clocks.max.graphics", "clocks.max.sm", "clocks.max.memory", "clocks.applications.graphics", "clocks.applications.mem",
    }

    for _, row := range records {
        name := fmt.Sprintf("%s", row[0])
        gpuidx := fmt.Sprintf("%s", row[1])
        gpuserial := fmt.Sprintf("%s", row[2])
        fmt.Println(row)
        for idx, value := range row[3:] {
            // process when value is int or float
            _,err := strconv.ParseFloat(value, 64)
            if err != nil {
                // nothing
            } else {
                metricName := strings.Replace(metricList[idx], ".", "_", -1)
                result := fmt.Sprintf("gpu_%s{gpu=\"%s\",idx=\"%s\",serial=\"%s\"} %s\n", metricName, name, gpuidx, gpuserial, value)
                fmt.Fprintf(response, result)
            }
        }
    }

}

func redirectMetrics(response http.ResponseWriter, request *http.Request) {
    http.Redirect(response, request, "/metrics", 301)
}

func main() {
    var addr string
    if len(os.Args) > 1 {
        addr = ":"+os.Args[1]
    } else {
        addr = ":9101"
    }
    fmt.Printf("Listen on %v\n", addr)
    http.HandleFunc("/", redirectMetrics )
    http.HandleFunc("/metrics/", metrics)
    err := http.ListenAndServe(addr, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}
