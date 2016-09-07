/*
 * Copyright © 2016 Jef Mathiot <jef@nonblocking.info>
 * This work is free. You can redistribute it and/or modify it under the
 * terms of the Do What The Fuck You Want To Public License, Version 2,
 * as published by Sam Hocevar. See the LICENSE.txt file for more details.
 */

package main

import(
  "bufio"
  "bytes"
  "encoding/csv"
  "fmt"
  "os/exec"
  "os"
  "regexp"
  "strconv"
  "time"
)

type Record struct {
  Rank int
  Domain string
  Support bool
}

func dig(domain string, attempts int) string {
  binary, lookupErr := exec.LookPath("dig")
  if lookupErr != nil {
      panic(lookupErr)
  }
  cmd := exec.Command(binary, "+dnssec", domain, "A")
  var out bytes.Buffer
  cmd.Stdout = &out
  err := cmd.Run()
  if err != nil {
    fmt.Println("Problem digging", domain, "(attempt ", attempts + 1, " of 10)")
    if attempts < 10 {
      time.Sleep(time.Second)
      return dig(domain, attempts + 1)
    } else {
      return ""
    }
  }
  return out.String()
}

func readCsv(path string) []Record {
  f, _ := os.Open(path)
  result, _ := csv.NewReader(bufio.NewReader(f)).ReadAll()
  result = result
  records := make([]Record, len(result))
  for i := range result {
    rank, _ := strconv.Atoi(result[i][0])
	  records[i] = Record{Rank: rank, Domain: result[i][1]}
  }
  return records
}

func rrsig(input string) bool{
  r, _ := regexp.Compile(`(?m)^\b((xn--)?[a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}\b\.\s+\d+\s+IN\s+RRSIG`)
  return r.MatchString(input)
}

func worker(id int, records <-chan Record, results chan<- Record) {
  for r := range records {
    support := rrsig(dig(r.Domain, 0))
    results <- Record{Rank: r.Rank, Domain: r.Domain,
                      Support: support}
  }
}

func writeToCsv(results []Record, path string) {
  os.Remove(path)
  f, _ := os.Create("result.csv")
  writer := csv.NewWriter(f)
  writer.Write([]string{"Rank", "Domain", "DNSSEC"})
  for _, r := range results {
    row := []string{strconv.Itoa(r.Rank), r.Domain, strconv.FormatBool(r.Support)}
    writer.Write(row)
  }
  writer.Flush()
  f.Close()
}

func printStats(results []Record) {
  total, supported := len(results), 0
  for j := range results {
    if results[j].Support {
      supported += 1
    }
  }
  rate := float64(supported) / float64(total) * 100
  fmt.Println("Total: ", total, ", supported: ",
              supported, "(" + strconv.FormatFloat(rate, 'f', -1, 64) + "%)")
}

func work(workers int){
  records := readCsv("top-1m.csv")
  results := make([]Record, len(records))

  jobs := make(chan Record, len(records))
  job_results := make(chan Record, len(records))
  for w := 0; w < workers; w++ {
    go worker(w, jobs, job_results)
  }

  for _, r := range records {
    jobs <- r
  }

  for j := range results {
    result := <-job_results
    results[j] = result
    fmt.Println("Processed domain #" + strconv.Itoa(j+1), results[j].Domain,
                ", DNSSEC: ", results[j].Support)
  }
  writeToCsv(results, "results.csv")
  printStats(results)
  close(jobs)
}

func main() {
  work(100)
}
