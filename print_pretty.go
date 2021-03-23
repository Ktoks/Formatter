package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

////// Helpers //////

type XNode struct {
	curNode string
	// curNodeName string
	children []XNode
	// do I add an int here to tell what kind of node it is???
	// could get rid of curNodeName
	// child depth int ??? this could slow it down, but paired with a hash table, this could also allow to go without recursion reading from root node
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// func handlenoIndent(noIndent bool, indent string, addindent string) (string, bool) {
// 	tempnum := 0
// 	if noIndent && indent != "" {
// 		tempnum = len(indent) - len(addindent)
// 		indent = strings.Repeat(" ", tempnum)
// 		noIndent = false
// 	}
// 	return indent, noIndent
// }

// tested, fine
func Clean(cleanedSlice []byte) string {
	var nodes string
	cleanedSlice = bytes.ReplaceAll(cleanedSlice, []byte("\n"), []byte(" "))
	cleanedSlice = bytes.ReplaceAll(cleanedSlice, []byte("\t"), []byte(" "))
	for i := 0; i < len(cleanedSlice); i++ {
		if (cleanedSlice[i]) == ' ' && cleanedSlice[i] == cleanedSlice[i-1] {
			continue
		} else {
			nodes += string(cleanedSlice[i])
		}
	}
	return nodes
}

////// Setup //////

func GetNodes(xmlSlice []byte) (XNode, int, int) {
	cleanedXml := Clean(xmlSlice)
	count := len(cleanedXml) ///////// git rid of after testing
	index := 0
	var rootNode XNode
	var totalNodes int // git rid of after testing
	beginning := 0
	for ; cleanedXml[index] != '>'; index++ {
		// fmt.Println(index)
		if cleanedXml[index] == '<' {
			beginning = index + 1
		}
	}
	rootNode.curNode = cleanedXml[beginning:index]

	rootNode, totalNodes = CreateNodeTree(rootNode, cleanedXml, index)
	// fmt.Printf("CurNodeName: %s\nchildren: %s\n", rootNode.curNode, rootNode.curNode)

	return rootNode, totalNodes, count // git rid of after testing
}

// fixed perl nodes
func CreateNodeTree(parentNode XNode, xmlStr string, index int) (XNode, int) {
	beginning := 0
	end := 0
	endNode := false

	for ; index < len(xmlStr); index++ {
		// begin node
		// fmt.Println(index)
		if xmlStr[index] == '<' {
			beginning = index
			end = 0
			// for xmlStr[index] == ' ' {
			// 	index++
			// }
			if xmlStr[index+1] == '/' || xmlStr[index+1] == '!' {
				endNode = true
			}

		} else if xmlStr[index] == '>' && index < len(xmlStr)-1 && xmlStr[index+1] != xmlStr[index] && xmlStr[index-1] != xmlStr[index] {
			// create new node to be put into it's parent
			var child XNode
			end = index

			// // child.curNodeName = nodeName // set nn
			// if firstSpace == 0 {
			// 	firstSpace = index - 1
			// } ////////////////////////////////////////////////////////////////////////////////
			child.curNode = xmlStr[beginning+1 : end]
			// end = 0
			if endNode {
				// this is the end of current parent node
				endNode = false
			} else { // if '>', not childless, and not end node
				// must be a parent node, recurse
				child, index = CreateNodeTree(child, xmlStr, index+1)
			}
			// put child into parent's children
			parentNode.children = append(parentNode.children, child)
		}

	}
	return parentNode, index
}

////// Main //////

func main() {

	inputFile := flag.String("i", "", "The full path to the input xml")
	// outputFile := flag.String("o", "", "The full path to the output xml")
	spaces := flag.String("s", "2", "The number of spaces preferred after each line, default is 2")
	flag.Parse()
	start := time.Now()

	// sOutputFile := *outputFile
	// if len(sOutputFile) == 0 {
	// 	sOutputFile = (*inputFile)[0:len(*inputFile)-4] + ".pretty.xml"
	// }

	myfile, err := os.ReadFile(*inputFile)
	if err != nil {
		fmt.Printf("Error reading input file:\n%v", err)
		log.Fatal(err)
	}
	// var data Data
	// xml.Unmarshal(myfile, &data)

	// fileInString := string(myfile)

	nodes, totalNodes, count := GetNodes(myfile)

	fmt.Printf("Compare: %s : %s\n", fmt.Sprintf("%d", count), fmt.Sprintf("%d", totalNodes))
	done := make(chan string, 30)
	fname := "./tmp/dat"
	for i := 1; i < 100; i++ {
		newname := fname + strconv.Itoa(i) + ".xml"
		if _, err := os.Stat(newname); err == nil {
			continue
		} else {
			fout, err := os.Create(newname)
			Check(err)

			/////////////////////////////
			// if false {
			// 	var newStr string
			// 	for j, node := range nodes {
			// 		newStr += node
			// 		j++
			// 	}
			// 	fout.WriteString(newStr)
			// }
			//////////////////////////////
			// first := true
			for sout := range Formatter(done, nodes, *spaces) {
				// if first {
				// 	// time.Sleep(2 * time.Second)
				// 	fout.WriteString(sout)
				// 	first = false
				// }
				fout.WriteString(sout)
			}

			//////////////////////////////
			// f.WriteString(nodes[0])
			// fmt.Println("made it: " + newname)
			break
		}
	}

	duration := time.Since(start)

	fmt.Printf("inputFile: %v\nduration: %v", *inputFile, duration)
}

////// Formatters //////

// writes out to file while worker puts things into the channel (concurrent)
func Formatter(done chan string, nodes XNode, spaces string) <-chan string {
	newl := "\n"
	nSpaces, err := strconv.Atoi(spaces)
	if err != nil {
		fmt.Printf("Error: Please input a proper integer!\n%v", err)
		log.Fatal(err)
	}
	indent := strings.Repeat(" ", nSpaces)
	addindent := strings.Repeat(" ", nSpaces)
	first := true
	nextAncestor := true
	extraIndent := false
	go Worker(done, nodes, indent, addindent, newl, first, nextAncestor, extraIndent)
	return done
}

// formats strings according to printPretty and puts them in the channel (Recursive and concurrent)
func Worker(out chan string, parentNode XNode, indent string, addindent string, newl string, first bool, nextAncestor bool, ExtraIndent bool) {
	index := 0

	if nextAncestor {
		indent += addindent
		nextAncestor = false
		ExtraIndent = true
	}

	// beginning of tag, index is 0
	if len(parentNode.curNode) > 0 {
		if parentNode.curNode[0] == '?' {
			// skip xml version
			Worker(out, parentNode.children[0], indent, addindent, newl, first, nextAncestor, ExtraIndent)
			return
		} else if parentNode.curNode[0] == '/' && len(indent) > 0 {
			tempnum := (len(indent) - len(addindent))
			indent = strings.Repeat(" ", tempnum)
			ExtraIndent = false
		} else if parentNode.curNode[0] == '!' {
			HandleSpecialNodes(out, parentNode, indent, newl)
			fmt.Println("!!!!!!!!", parentNode.curNode)
			// instead of return, set a flag for !?
			return
		}
	}
	if first {
		defer close(out)
		first = false
		indent = ""
	}

	// opening tag
	out <- indent + "<"

	// moreThanOne := false
	for ; index < len(parentNode.curNode); index++ {
		if parentNode.curNode[index] == ' ' {
			// moreThanOne = true
			// indent, noIndent = handlenoIndent(noIndent, indent, addindent)

			if index+1 == len(parentNode.curNode) {
				out <- newl
			} else {
				out <- newl + indent
			}
		} else {
			out <- string(parentNode.curNode[index])
		}
	}

	index = len(parentNode.curNode)
	// closing tag // I think this block's finished
	if len(parentNode.children) > 0 {

		if ExtraIndent && parentNode.curNode[index-1] != '/' {
			nextAncestor = true
			ExtraIndent = false
			out <- newl + indent + ">" + newl
		} else {
			out <- fmt.Sprintf(">%s", newl)
		}
		for j, child := range parentNode.children {
			Worker(out, child, indent, addindent, newl, first, nextAncestor, ExtraIndent)
			j++
		}
	}
	out <- fmt.Sprintf(">%s", newl)
	// }
}

func HandleSpecialNodes(out chan string, node XNode, indent string, newl string) {
	if strings.Contains(node.curNode, "%%") {
		cleanNode := strings.TrimSpace(node.curNode)
		output := fmt.Sprintf("%s<%s%s", indent, cleanNode, newl)
		out <- output
	} else if strings.Contains(node.curNode, "?") {
	} else if node.curNode[0:3] == "!--" {
		out <- "<" + node.curNode + ">" + newl

	} else if strings.Contains(node.curNode, "CDATA") {
		out <- fmt.Sprintf("%s<![CDATA[", indent)

		temp := strings.SplitAfter(node.curNode, "[CDATA[")
		temp = strings.Split(temp[1], "]")

		PerlCode := FormatPerl(temp, newl) // returns lines of perl in a slice from perltidy

		fmt.Println("???????", string(PerlCode))

		out <- fmt.Sprintf("%s%s]]>%s", PerlCode, indent, newl)
	} else {
		fmt.Println("ERROR: Unexpected node type. " + node.curNode)
	}
}

////// Formatter for Perl //////

func FormatPerl(cdata []string, newl string) []byte {
	// make a temp dir
	var formattedData []byte
	var strippedData string

	tempFile, err := ioutil.TempFile("./tmp/.", "output")
	if err != nil {
		log.Fatal(err)
	}
	// defer os.Remove(tempFile.Name())

	// split the perl by \n
	for i, data := range cdata {
		strippedData += strings.TrimSpace(data)
		i++
	}

	_, err = tempFile.WriteString(strippedData)
	if err != nil {
		fmt.Println("Error: ", err)
		log.Fatal(err)
	}
	newTempFile := tempFile.Name()

	cmd := "perltidy"
	args := fmt.Sprintf("./tmp/%s -b -l=0", newTempFile)
	// fmt.Println("Command output: ", cmd, args)
	com := exec.Command(cmd, args)

	// var NeedsNewl bool
	if com.ProcessState.ExitCode() != 0 {
		fmt.Println("e: ", com.ProcessState.ExitCode())
		// fmt.Printf("WARNING: Invalid perl code found... Writing out CDATA as is. see: %s\n", strippedData)
		// NeedsNewl = true
	}

	formattedData, err = os.ReadFile(newTempFile)
	if err != nil {
		log.Fatal(err)
	}
	// for i, dis := range After {

	// 	i++
	// }
	// if NeedsNewl {
	// 	formattedData[len(formattedData)-1] += '\n'
	// }
	return formattedData
}
