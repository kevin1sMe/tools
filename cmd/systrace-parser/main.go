package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/alecthomas/kingpin.v2"
)

//go:embed VERSION
var version string

var (
	filename = kingpin.Flag("filename", "filename of Chrome's bookmarks").Short('f').String()
	output   = kingpin.Flag("output", "output filename, append '.out' if empty").Short('o').String()

	input_dir  = kingpin.Flag("input_dir", "input directory").Short('i').String()
	output_dir = kingpin.Flag("output_dir", "output directory").Short('d').String()
)

func main() {
	// app.HelpFlag.Short('h')
	kingpin.Version(version)
	kingpin.Parse()

	// 单文件转换模式
	if *filename != "" {
		ParseSysTrace(*filename, *output)
	}

	// 目录模式
	if *input_dir != "" {
		res := TraverseDir(*input_dir, ".html")

		log.Println("============")
		log.Println(res)
		for _, r := range res {
			dest := OutputFilename(r, *input_dir, *output_dir)
			ParseSysTrace(r, dest)
		}
	}
}

func OutputFilename(filename, inputDir, outputDir string) string {
	if inputDir == "." {
		inputDir = "." + string(os.PathSeparator)
	}

	if outputDir == "" {
		outputDir = inputDir
	}

	if outputDir[len(outputDir)-1] != os.PathSeparator {
		outputDir += string(os.PathSeparator)
	}

	f := strings.Replace(filename, inputDir, outputDir, -1)

	filenameAll := path.Base(f)
	fileExt := path.Ext(f)
	return fmt.Sprintf("%s/%s%s", path.Dir(f), filenameAll[0:len(filenameAll)-len(fileExt)], ".txt")
}

// 解析一个systrace html文件并输出到指定位置
// </script>
// <!-- BEGIN TRACE -->
//   <script class="trace-data" type="application/text">
func ParseSysTrace(src, dest string) error {
	if dest == "" {
		dest = src + ".out"
	} else {
		os.MkdirAll(path.Dir(dest), os.ModePerm)
	}

	fmt.Printf("src:%s dest:%s\n", src, dest)

	// 读取原始文件
	f, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	dom, err := goquery.NewDocumentFromReader(bufio.NewReader(f))
	if err != nil {
		log.Fatalln(err)
	}

	dom.Find(".trace-data").Each(func(i int, s *goquery.Selection) {
		fmt.Println("find trace-data")
		// res, _ := s.Html()
		res := s.Text()
		fmt.Printf("file: %s html size:%d\n", src, len(res))

		//将结果写文件
		err = os.WriteFile(dest, []byte(res), 0644)
		if err != nil {
			log.Println("write output failed, ", err)
		}

	})
	return err
}

func TraverseDir(srcDir string, filterExt string) (results []string) {
	dir, err := ioutil.ReadDir(srcDir)
	if err != nil {
		log.Fatalln(err)
	}

	for _, fi := range dir {
		filename := fmt.Sprintf("%s%s%s", srcDir, string(os.PathSeparator), fi.Name())
		if fi.IsDir() { // 目录
			results = append(results, TraverseDir(filename, filterExt)...)
		} else {
			log.Printf("文件：%s", filename)
			fileSuffix := path.Ext(filename)
			if fileSuffix != filterExt {
				log.Printf("not match, skip 文件：%s", filename)
			} else {
				log.Printf("match, 文件：%s", filename)
				results = append(results, filename)
			}
		}
	}
	return
}
