package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/go-sql-driver/mysql"
	"github.com/greg0/go-mysqldump"
	"gopkg.in/yaml.v2"
)

const (
	// Info messages
	Info = 1 << iota // a == 1 (iota has been reset)

	// Warning Messages
	Warning = 1 << iota // b == 2

	// Error Messages
	Error = 1 << iota // c == 4

	cName = "dump"

	cAddr = "127.0.0.1:3306"

	cUsr = "root"

	cPwd = "root"
)

type ConnectionConfig struct {
	ConnName string `yaml:"name"`
	Address  string `yaml:"address"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"dbname"`
}

// Options model for commandline arguments
type Options struct {
	ConnName            string
	Address             string
	UserName            string
	Password            string
	Database            string
	IgnoredTables       []string
	StructureOnlyTables []string
	OutputDirectory     string
}

func main() {

	options := GetOptions()

	// Open connection to database
	config := mysql.NewConfig()
	config.User = options.UserName
	config.Passwd = options.Password
	config.DBName = options.Database
	config.Net = "tcp"
	config.Addr = options.Address

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		fmt.Println("Error opening database: ", err)
		return
	}

	dumpFilenameFormat := fmt.Sprintf("%s-%s-2006-01-02T15:04:05", options.ConnName, config.DBName) // accepts time layout string and add .sql at the end of file
	// Register database with mysqldump
	dumper, err := mysqldump.Register(
		db,
		options.OutputDirectory,
		dumpFilenameFormat,
		options.StructureOnlyTables,
		options.IgnoredTables,
	)
	if err != nil {
		fmt.Println("Error registering databse:", err)
		return
	}

	// Dump database to file
	err = dumper.Dump()
	if err != nil {
		fmt.Println("Error dumping:", err)
		fmt.Println("Removing dump file")
		os.Remove(dumper.File.Name())
		return
	}
	fmt.Printf("File is saved to %s/%s", options.OutputDirectory, dumpFilenameFormat)

	// Close dumper, connected database and file stream.
	dumper.Close()
}

func GetOptions() *Options {

	var connection string
	flag.StringVar(&connection, "connection", "", "Yaml config with connection parameters. Overrides flag arguments. Optional")

	var connName string
	flag.StringVar(&connName, "name", cName, "Dump name. Default 'dump' ")

	var address string
	flag.StringVar(&address, "addr", cAddr, "Database address host:port")

	var username string
	flag.StringVar(&username, "user", cUsr, "Database username")

	var password string
	flag.StringVar(&password, "pass", cPwd, "Database password")

	var database string
	flag.StringVar(&database, "dbname", "", "Database name")

	var ignoredTables string
	flag.StringVar(&ignoredTables, "ignore", "", "File path containing list of ignored tables. Optional. File can contains regex expressions. Each expresion in new line")

	var structOnlyTables string
	flag.StringVar(&structOnlyTables, "structOnly", "", "File path containing list of ignored tables. Optional. File can contains regex expressions. Each expresion in new line")

	var outputdir string
	flag.StringVar(&outputdir, "output", "", "Dump output dir. Default is current working directory")

	if len(os.Args) <= 1 || os.Args[1] == "--help" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	flag.Parse()

	if connection != "" {
		var c ConnectionConfig
		c.loadFile(connection)

		connName = c.ConnName
		address = c.Address
		username = c.UserName
		password = c.Password
		database = c.Database
	}

	if outputdir == "" {
		dir, err := os.Getwd()
		checkErr(err)
		outputdir = dir
	}

	if database == "" {
		printMessage("Database name can't be empty", Error)
		os.Exit(1)
	}

	ignoredTablesArray := []string{}
	if ignoredTables != "" {
		ignoredTablesArray = FillArrayWithFileLines(ignoredTables)
	}

	structOnlyTablesArray := []string{}
	if structOnlyTables != "" {
		structOnlyTablesArray = FillArrayWithFileLines(structOnlyTables)
	}

	var o Options
	opts := o.create(
		connName,
		address,
		username,
		password,
		database,
		ignoredTablesArray,
		structOnlyTablesArray,
		outputdir)

	stropts, _ := json.MarshalIndent(opts, "", "\t")
	printMessage("Running with parameters", Info)
	printMessage(string(stropts), Info)
	printMessage("Running on operating system : "+runtime.GOOS, Info)

	return opts
}

func FillArrayWithFileLines(filePath string) []string {
	file, err := os.Open(filePath)
	checkErr(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var array []string
	for scanner.Scan() {
		array = append(array, scanner.Text())
	}

	return array
}

func (o *Options) create(name string, address string, username string, password string, database string, ignoredTables []string, structOnlyTables []string, outputDir string) *Options {

	database = strings.Replace(database, " ", "", -1)
	database = strings.Replace(database, " , ", ",", -1)
	database = strings.Replace(database, ", ", ",", -1)
	database = strings.Replace(database, " ,", ",", -1)

	return &Options{
		ConnName:            name,
		Address:             address,
		UserName:            username,
		Password:            password,
		Database:            database,
		IgnoredTables:       ignoredTables,
		StructureOnlyTables: structOnlyTables,
		OutputDirectory:     outputDir,
	}
}

func printMessage(message string, messageType int) {
	colors := map[int]color.Attribute{Info: color.FgGreen, Warning: color.FgHiYellow, Error: color.FgHiRed}
	color.Set(colors[messageType])
	fmt.Println(message)
	color.Unset()
}

func checkErr(err error) {
	if err != nil {
		color.Set(color.FgHiRed)
		panic(err)
		color.Unset()
	}
}

func (c *ConnectionConfig) loadFile(filePath string) *ConnectionConfig {
	c.ConnName = cName
	c.Address = cAddr
	c.UserName = cUsr
	c.Password = cPwd

	yamlFile, err := ioutil.ReadFile(filePath)
	checkErr(err)
	err = yaml.Unmarshal(yamlFile, c)
	checkErr(err)

	return c
}
