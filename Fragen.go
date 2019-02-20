package main

import (
	"bytes"
	"encoding/base64"
   "encoding/json"
	"flag"
	"html/template"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/jpeg"
	"image/png"
   "io"
	"log"
   "fmt"
	"math/cmplx"
	"net/http"
   "os"
   "reflect"
   "runtime"
	"strconv"
   "strings"
   "time"
)

type Config struct {
  Scale float64
  Width  int
  Height int
  MinEdge int
  X0 float64
  Y0 float64
  Parall int
  Axes bool
  Max_iteration int
  CentX int
  CentY int
}

var cfg Config = Config{}

var PalCust []color.Color
	
var verbose bool = false

var root = flag.String("root", ".", "file system path")

func initPal() {
  PalCust = make([]color.Color, 256) 
  for n := 0; n < 256; n++ {
     g := uint8(32 * ((n ) & 0x7))
     b := uint8(32 * ((n >> 3) & 0x7))
     r := uint8(64 * ((n >> 6) & 0x3))
     PalCust[n] = color.RGBA{r, g, b, 0xff}        
  }
}

func main() {
  cfgFile := "Fragen.json"
  file, _ := os.Open(cfgFile)
  defer file.Close()
  decoder := json.NewDecoder(file)
//  cfg := Config{}
  err := decoder.Decode(&cfg)
  if err != nil || cfg.Scale == 0 {
    if err!= nil { fmt.Println("error:", err); }
    fmt.Println("setting default config")
    cfg.Scale = 2.2
    cfg.Width = 256 
    cfg.Height = 256 
    cfg.X0 = 1.0
    cfg.Y0 = 0.0
    cfg.Axes = true
    cfg.Max_iteration = 512
    cfg.Parall = 16
    cfg.CentX = 128
    cfg.CentY = 128
  } else {
   if cfg.Width  == 0 { cfg.Width = 128 }
   if cfg.Height == 0 { cfg.Height = 128 }
  }
  fmt.Println("config:", cfg)


 	initPal()
//	log.Println("palette\n:", PalCust)
//   var srv http.Server

	http.HandleFunc("/blue/", blueHandler)
	http.HandleFunc("/red/",  redHandler)
	http.HandleFunc("/frac/", fracHandler)
	http.Handle("/dir/", http.FileServer(http.Dir(*root)))
	http.HandleFunc("/", defaultHandler)
   port := "8081"
   log.Println("Listening on ", port)
//   log.Println("srv=", srv)
	err = http.ListenAndServe(":" + port, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func blueHandler(w http.ResponseWriter, r *http.Request) {
	m := image.NewRGBA(image.Rect(0, 0, 240, 240))
	blue := color.RGBA{0, 0, 255, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

	var img image.Image = m
	writeImage(w, &img)
}

func redHandler(w http.ResponseWriter, r *http.Request) {
	m := image.NewRGBA(image.Rect(0, 0, 240, 240))
	blue := color.RGBA{255, 0, 0, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)

	var img image.Image = m
	writeImageWithTemplate(w, &img, ImageTemplate)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	m := image.NewRGBA(image.Rect(0, 0, 128, 128))
	green := color.RGBA{0, 255, 0, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{green}, image.ZP, draw.Src)

//	var img image.Image = m
   e := reflect.ValueOf(http.DefaultServeMux).Elem()
   log.Println("defMux=", e)

   buffer := "<html><body><a href='/red'>red</a><br /><a href='/blue'>blue</a><br /><a href='/dir'>dir</a><br /><a href='f/rac'>frac</a><br /></body></html>"

	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer)))
   io.WriteString(w, buffer)
//	if _, err := w.Write(buffer); err != nil {
//		log.Println("unable to write page.")
//	}
     
}

func SaveConfig() {
  cfgFile := "Fragen.json"
  file, _ := os.Create(cfgFile)
  defer file.Close()
  enc := json.NewEncoder(file)
  enc.SetIndent("", "  ")
  enc.Encode(&cfg)
}

var imarr [][]uint8
var duratn string
var method int
var selPal string

func fracHandler(w http.ResponseWriter, r *http.Request) {
   r.ParseForm()
   if (verbose) { fmt.Println("fracHandler") }
//   for v := r {
   if (verbose) {     fmt.Println(r.Form) }
//        fmt.Println("key:", v)
//   }
   if (verbose) { 
     fmt.Println("\nForm inputs:")
     for k, v := range r.Form {
        fmt.Printf("%10s = %s\n", k, strings.Join(v, "") )
//        fmt.Println("val:", strings.Join(v, ""))
     }
   }

   // when called first time use read config
   actn := r.Form.Get("actName")
   fmt.Printf("##%s## %d\n", actn, len(actn))
   if actn != "" && len(actn) > 0 {
     fmt.Println("reading form values")

     var err error
     var vf64 float64
     var vi int
     vf64, err = strconv.ParseFloat(r.Form.Get("scale"), 64)
     if err == nil && vf64 > 0 { cfg.Scale = vf64 }

     vi,  err = strconv.Atoi(r.Form.Get("width"))
     if err == nil && vi > 0 { cfg.Width = vi }
 
     vi, err = strconv.Atoi(r.Form.Get("height"))
     if err == nil && vi > 0 { cfg.Height = vi }

     vf64, err = strconv.ParseFloat(r.Form.Get("x0"), 64)
     if err == nil { cfg.X0 = vf64 }

     vf64, err = strconv.ParseFloat(r.Form.Get("y0"), 64)
     if err == nil { cfg.Y0 = vf64 }

     vi, err = strconv.Atoi(r.Form.Get("centerX"))
     if err == nil && vi > 0 { cfg.CentX = vi }

     vi, err = strconv.Atoi(r.Form.Get("centerY"))
     if (err == nil) && vi > 0 { cfg.CentY = vi }

     vi, err = strconv.Atoi(r.Form.Get("paralel"))
     if (err == nil) && vi > 0 { cfg.Parall = vi }

     vi, err = strconv.Atoi(r.Form.Get("maxIter"))
     if (err == nil) && vi > 0 { cfg.Max_iteration = vi }

     if "on" == r.Form.Get("showAxes") { cfg.Axes = true } else { cfg.Axes = false }

     if "on" == r.Form.Get("altMethod") { method = 2 } else { method = 1 }
   
     selPal = r.Form.Get("selPal")
     if (selPal == "") { selPal = "1" }

   } else {
     fmt.Println("cfg =", cfg)
   }

   if (verbose) { log.Println("axes = ", cfg.Axes, ", method = ", method) }

   cfg.MinEdge = cfg.Width
   if (cfg.MinEdge > cfg.Height) { cfg.MinEdge = cfg.Height }
   
   if (actn == "zoomin")  { cfg.Scale /= 1.1 }
   if (actn == "zoomout") { cfg.Scale *= 1.1 }
   if (actn == "wadd")    { cfg.Width += 16 }
   if (actn == "wsub")    { cfg.Width -= 16 }
   if (actn == "hadd")    { cfg.Height += 16 }
   if (actn == "hsub")    { cfg.Height -= 16 }
   if (actn == "xadd")    { cfg.X0 += 15.0 * cfg.Scale / float64(cfg.MinEdge) }
   if (actn == "xsub")    { cfg.X0 -= 15.0 * cfg.Scale / float64(cfg.MinEdge) }
   if (actn == "yadd")    { cfg.Y0 += 15.0 * cfg.Scale / float64(cfg.MinEdge) }
   if (actn == "ysub")    { cfg.Y0 -= 15.0 * cfg.Scale / float64(cfg.MinEdge) }
   if (actn == "frimg")   { 
      cfg.X0 -= float64(cfg.CentX - cfg.Width  / 2) * cfg.Scale / float64(cfg.MinEdge) 
      cfg.Y0 -= float64(cfg.CentY - cfg.Height / 2) * cfg.Scale / float64(cfg.MinEdge) 
   }

//	m := image.NewRGBA(image.Rect(0, 0, width, height))
   if cfg.Width  == 0 { cfg.Width = 128 }
   if cfg.Height == 0 { cfg.Height = 128 }


   if (verbose) { log.Print("START") }
	start := time.Now()
	//	log.Print("Size: ", *width, *height)
	//	X0 = int(0.8 * float64(*width))
	//	Y0 = int(*height / 2)
//	scale = sc
   if (verbose) { log.Print("X0, Y0, scale  ", cfg.X0, cfg.X0, cfg.Scale) }
	// var imarr [*width][*height]uint8
	imarr = make([][]uint8, cfg.Height)
	for i := range imarr {
		imarr[i] = make([]uint8, cfg.Width)
	}

	// Create a colored image of the given width and height.
	// img := image.NewNRGBA(image.Rect(0, 0, *width, *height))
   //selPal := 2
   var m *image.Paletted
   if selPal == "1"        { m = image.NewPaletted(image.Rect(0, 0, cfg.Width, cfg.Height), palette.Plan9) 
   } else if selPal == "2" { m = image.NewPaletted(image.Rect(0, 0, cfg.Width, cfg.Height), palette.WebSafe) 
   } else                  { m = image.NewPaletted(image.Rect(0, 0, cfg.Width, cfg.Height), PalCust) }

	// var nx, ny float64
	if method == 1 {
		calcFrac1()
	} else {
		calcFrac2()
	}

	for y := 0; y < cfg.Height; y++ {
		for x := 0; x < cfg.Width; x++ {
  		  if cfg.Axes && ((x == cfg.Width / 2) || (y == cfg.Height  / 2)) {
  		     m.SetColorIndex(x, y, 255 - imarr[y][x])
        } else {
  			  m.SetColorIndex(x, y, imarr[y][x])
        }
		}
	}

//	green := color.RGBA{0, 250, 0, 255}
//	draw.Draw(m, m.Bounds(), &image.Uniform{green}, image.ZP, draw.Src)
//	for x := 0; x < width; x++ {
//      m.Set(x, x * height / width, color.RGBA{0, 0, 0, 255})
//   }

   fmt.Printf("x0=%f, y0=%f, scale=%f\n", cfg.X0, cfg.Y0, cfg.Scale)
   duratn = time.Now().Sub(start).String()
   SaveConfig()
	var img image.Image = m
 	writeImageWithTemplate(w, &img, "FormTemplate")
}

var ImageTemplate string = `<!DOCTYPE html>
<html lang="en"><head></head>
<body><div>{{.Title}}</div>
<img src="data:image/jpg;base64,{{.Image}}"></body>`

func getChecked(b bool) string {
	ret := ""
	if b { ret = "checked"}
   if (verbose) { log.Println("getChecked(", b, ") = ", ret) }
	return ret
}

type Option struct {
	Val, Txt string
	Sel  bool
}

// Writeimagewithtemplate encodes an image 'img' in jpeg format and writes it into ResponseWriter using a template.
func writeImageWithTemplate(w http.ResponseWriter, img *image.Image, templStr string) {

	buffer := new(bytes.Buffer)
//	if err := jpeg.Encode(buffer, *img, nil); err != nil {
	if err := png.Encode(buffer, *img); err != nil {
		log.Fatalln("unable to encode image.")
	}

	str := base64.StdEncoding.EncodeToString(buffer.Bytes())
	htmlTemplate := "FragenTemplate.html"
   tmpl, err := template.ParseFiles(htmlTemplate)
   //tmpl, err := template.New("image").Parse(templStr)
	if err != nil {
		log.Println("unable to parse template.")
      log.Println(htmlTemplate)
	} else {
      fmt.Printf("x0=%f, y0=%f, scale=%f, width=%d, height=%d\n", cfg.X0, cfg.Y0, cfg.Scale, cfg.Width, cfg.Height)
		data := map[string]interface{} {
              "Image": str, 
              "Title":"Fragen", 
              "scale":cfg.Scale, 
              "width":cfg.Width, 
              "height":cfg.Height, 
              "x0":cfg.X0, 
              "y0":cfg.Y0,
			  "log":duratn,
			  "paralel":cfg.Parall,
			  "maxIter":cfg.Max_iteration, 
			  "showAxes":getChecked(cfg.Axes),
			  "altMethod":getChecked(method == 2),
			  "selPal":selPal,
			  "palletes":[]Option{Option{"1", "Plan9", selPal=="1"}, 
			                      Option{"2", "WebSafe", selPal == "2"},
			                      Option{"3", "Custom", selPal=="3"},
			                      },
			 }
		
	   if (verbose) {
         log.Println("data = ") 
		   for k, v := range data {	 
		     if (k != "Image") { log.Println(k, v) }	
		   } 
      }
		if err = tmpl.Execute(w, data); err != nil {
			log.Println("unable to execute template.")
		}
	}
}

func check(z0 complex128) int {
	z := complex(0, 0)

	iteration := 0
	for iteration = 0; cmplx.Abs(z) < 4 && iteration < cfg.Max_iteration; iteration++ {
		z = z*z + z0
	}
	return iteration
}

func process(x int, y int, line []uint8, m chan bool) {
//    log.Print(y)
	var nx, ny float64
   var ite int
   ny = cfg.Scale / float64(cfg.MinEdge) * float64(y - cfg.Height / 2) - cfg.Y0 
	for x := 0; x < cfg.Width; x++ {
 	   nx = cfg.Scale / float64(cfg.MinEdge) * float64(x - cfg.Width / 2) - cfg.X0
      ite = check(complex128(complex(nx, ny)))
		line[x] = uint8(ite & 255)
	}
	m <- true
}

func calcFrac1() {
 	runtime.GOMAXPROCS(cfg.Parall)

   if (verbose) { log.Println("calcFrac1()") }
   fmt.Printf("w=%d, h=%d, it=%d, mined=%d\n", cfg.Width, cfg.Height, cfg.Max_iteration, cfg.MinEdge)
   fmt.Printf("rect(%f, %f, %f, %f)\n", 
          cfg.Scale * float64(0 - cfg.Width  / 2) - cfg.X0, 
          cfg.Scale * float64(0 - cfg.Height / 2) - cfg.Y0, 
          cfg.Scale * float64(cfg.Width  - cfg.Width / 2)  - cfg.X0, 
          cfg.Scale * float64(cfg.Height - cfg.Height / 2) - cfg.Y0 )
	mb := make(chan bool, cfg.Height)
	for y := 0; y < cfg.Height; y++ {
		go process(0, y, imarr[y], mb)
	}

	for y := 0; y < cfg.Height; y++ {
		<-mb
	}
   log.Println("calcFrac1() - end")
}

func calcFrac2() {
   if (verbose) { log.Println("calcFrac2()") }
	mb := make(chan bool, cfg.Parall)
	p := 0
	for y := 0; y < cfg.Height; y++ {
		go process(0, y, imarr[y], mb)
		if p >= cfg.Parall {
			<-mb
		} else {
			p++
		}
	}

	for ; p > 0; p-- {
		<-mb
	}
}


// writeImage encodes an image 'img' in jpeg format and writes it into ResponseWriter.
func writeImage(w http.ResponseWriter, img *image.Image) {

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, *img, nil); err != nil {
		log.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
}
