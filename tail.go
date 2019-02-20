package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "flag"
)

func main() {
    var fileName = "::::::::"
    count := flag.Int("n", 5, "display at least n last lines from file")
    line_length := flag.Int("len", 32, "line length")
    verbose := flag.Bool("v", false, "verbose")
//    _ = flag.Bool("h", false, "help")

    flag.Parse()

//    fmt.Println(flag.Args(), len(flag.Args()), flag.Args()[0])

    if len(os.Args) <= 1 {
      fmt.Println("Missing argument ")
      return    
    } else {
      if len(flag.Args()) > 0 {
         fileName = flag.Args()[0]
      }
    }
   
    if fileName == "::::::::" {
       fmt.Println("Missing filename")
       return
    }
//    var stat, err = os.Stat(fileName) 
//    if err != nil {
//        log.Fatal(err)
//    }
//    var fsize = stat.Size()
//    fmt.Println("file size =", fsize)
    
//    fmt.Println("open " + fileName)    

    file, err := os.Open(fileName)
    if err != nil {
        log.Fatal(err)
    }

    defer file.Close()

    var nb int64 = int64(*line_length)
    var nc = 0
    var step int64 = 0
    for nc < *count {
      step += 1
      file.Seek(-step * nb, 2)

      scanner := bufio.NewScanner(file)
      nc = 0
      for scanner.Scan() {
        nc += 1
      }
      if err := scanner.Err(); err != nil {
          log.Fatal(err)
      }

    }
    if *verbose {
       fmt.Println("step =", step, " lines = ", nc, "bytes =", -step * nb)
    }

    file.Seek(-step * nb, 2)
    scanner := bufio.NewScanner(file)

    var firstLine = false
    for scanner.Scan()  {
        if nc <= *count && firstLine {
            fmt.Println(scanner.Text())
        }
        firstLine = true
        nc -= 1
    }
   
}
