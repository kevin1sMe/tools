package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"gopkg.in/alecthomas/kingpin.v2"
)

//go:embed VERSION
var version string

var (
	filename = kingpin.Flag("filename", "filename of Chrome's bookmarks").Short('f').String()
	output   = kingpin.Flag("output", "output filename, append '.out' if empty").String()

	input_dir  = kingpin.Flag("input_dir", "input directory").Short('i').String()
	output_dir = kingpin.Flag("output_dir", "output directory").Short('o').String()
)

func main() {
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version)
	kingpin.Parse()

	// 单文件转换模式
	if *filename != "" {
		ParseSysTrace(*filename, *output)
	}

	// 目录模式
	if *input_dir != "" {
		srcDir := filepath.Dir(*input_dir + string(os.PathSeparator) + "dump")
		if *output_dir == "" { // 如果没有指定输出目录，使用input目录
			*output_dir = *input_dir
		}

		destDir := filepath.Dir(*output_dir + string(os.PathSeparator) + "dump")
		log.Printf("trim dir ==> srcDir:%s, destDir:%s\n", srcDir, destDir)

		// 遍历目录获取符合过滤条件的文件列表
		res := TraverseDir(srcDir, ".html")
		log.Println("-----------file list found begin ------")
		log.Println(res)
		log.Println("-----------file list found end   ------")
		for _, r := range res {
			dest := OutputFilename(r, srcDir, destDir)
			ParseSysTrace(r, dest)
		}
	}
}

func OutputFilename(filename, inputDir, outputDir string) string {
	f := strings.Replace(filename, inputDir, outputDir, 1)
	log.Println("after => filename:", filename, ", inputDir:", inputDir, ", outputDir:", outputDir, "f:", f)

	filenameAll := filepath.Base(f)
	fileExt := filepath.Ext(f)
	output := fmt.Sprintf("%s%s%s%s", filepath.Dir(f), string(os.PathSeparator), filenameAll[0:len(filenameAll)-len(fileExt)], ".txt")
	log.Println("after => path:", filepath.Dir(f), filenameAll, fileExt, ", output: ", output)
	return output
}

// 解析一个systrace html文件并输出到指定位置
// </script>
// <!-- BEGIN TRACE -->
//   <script class="trace-data" type="application/text">
func ParseSysTrace(src, dest string) error {
	if dest == "" {
		dest = src + ".out"
	} else {
		if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
			log.Fatalln(err)
		}
	}

	fmt.Printf("\nsrc:%s dest:%s\n", src, dest)

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
		// fmt.Println("find trace-data")
		// res, _ := s.Html()
		res := s.Text()
		fmt.Printf("file: %s html size:%d\n", src, len(res))

		// 去掉第一个空行
		trimNewline := strings.Replace(res, "\n", "", 1)
		//将结果写文件
		err = os.WriteFile(dest, []byte(trimNewline), 0644)
		if err != nil {
			log.Println("write output failed, ", err)
		}
		fmt.Printf("[SUCC] create file: %s\n", dest)

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
			fileSuffix := filepath.Ext(filename)
			if fileSuffix != filterExt {
				log.Printf("not match, skip file:%s", filename)
			} else {
				log.Printf("match, file: %s", filename)
				results = append(results, filename)
			}
		}
	}
	return
}
