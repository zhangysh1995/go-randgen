package main

import (
	"fmt"
	"github.com/pingcap/go-randgen/gendata"
	"github.com/pingcap/go-randgen/grammar"
	"github.com/pingcap/go-randgen/grammar/sql_generator"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)


var zzPath string
var yyPath string

var queries int
var maxRecursive int
var root string

var debug bool
var skipZz bool
var seed int
var outPath string

var rootCmd = &cobra.Command{
	Use:   "go-randgen",
	Short: "QA tool for fuzzy test just like mysql go-randgen",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		rand.Seed(int64(seed))
		return nil
	},
}

// init command flag
func initCmd() {
	rootCmd.PersistentFlags().StringVarP(&zzPath, "zz", "Z",
		"", "zz file path, go go-randgen have a default zz")
	rootCmd.PersistentFlags().StringVarP(&yyPath, "yy", "Y","", "yy file path, required")
	rootCmd.PersistentFlags().IntVarP(&queries, "queries", "Q", 100, "random sql num generated by zz, if it is negative(like -1), exec subcommand will generate endless sql")
	rootCmd.PersistentFlags().StringVarP(&root, "root", "R", "query", "root bnf expression to generate sqls")

	rootCmd.PersistentFlags().IntVar(&maxRecursive, "maxrecur", 5,
		"yy expression most recursive number, if you want recursive without limit ,set it <= 0")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false,
		"print detail generate path")
	rootCmd.PersistentFlags().BoolVar(&skipZz, "skip-zz", false,
		"skip gen data phase, only use yy to generate random sqls")
	rootCmd.PersistentFlags().IntVar(&seed, "seed", time.Now().Nanosecond(),
		"random number seed, default time.Now().Nanosecond()")
	rootCmd.PersistentFlags().StringVarP(&outPath, "output", "O","output",
		"sql output file path")

	rootCmd.AddCommand(newExecCmd())
	rootCmd.AddCommand(newGentestCmd())
	rootCmd.AddCommand(newGenDataCmd())
	rootCmd.AddCommand(newGensqlCmd())
	rootCmd.AddCommand(newZzCmd())
}

func main() {
	initCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(rootCmd.UsageString())
		os.Exit(1)
	}
}

func getDdls() ([]string, gendata.Keyfun) {
	var zzBs []byte
	var err error
	if zzPath != "" {
		log.Printf("load zz from %s\n", zzPath)
		zzBs, err = ioutil.ReadFile(zzPath)
		if err != nil {
			log.Fatalf("load zz fail, %v\n", err)
		}
	} else {
		log.Println("load default zz")
	}

	zz := string(zzBs)

	ddls, keyf, err := gendata.ByZz(zz)
	if err != nil {
		log.Fatalln(err)
	}

	return ddls, keyf
}

func loadYy() string {
	log.Printf("load yy from %s\n", yyPath)
	yyBs, err := ioutil.ReadFile(yyPath)
	if err != nil {
		log.Fatalf("Fatal Error: load yy from %s fail, %v\n", yyPath, err)
	}

	yy := string(yyBs)

	return yy
}

func getRandSqls(keyf gendata.Keyfun) []string {

	yy := loadYy()

	randomSqls, err := grammar.ByYy(yy, queries, root, maxRecursive, keyf, debug)
	if err != nil {
		log.Fatalln("Fatal Error: " + err.Error())
	}

	return randomSqls
}

func getIter(keyf gendata.Keyfun) sql_generator.SQLIterator {
	yy := loadYy()

	iterator, err := grammar.NewIter(yy, root, maxRecursive, keyf, debug)
	if err != nil {
		log.Fatalln("Fatal Error: " + err.Error())
	}
	return iterator
}

func dumpRandSqls(sqls []string) {
	path := outPath+".rand.sql"
	err := ioutil.WriteFile(path,
		[]byte(strings.Join(sqls, ";\n") + ";"), os.ModePerm)
	if err != nil {
		log.Printf("write random sql in dist fail, %v\n", err)
	}

	log.Printf("dump random sqls in %s ok\n", path)
}