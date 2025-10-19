package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Field struct {
    Key   string
    Value string
}

type JobBlock struct {
    Fields []Field
}

const PREFIX = "DV03_EXA-"

func modifyBlocks(blocks []JobBlock) {
    for i := range blocks {
        for j := range blocks[i].Fields {
            if blocks[i].Fields[j].Key == "insert_job" {
                blocks[i].Fields[j].Value = PREFIX + blocks[i].Fields[j].Value
            }
		    if blocks[i].Fields[j].Key == "box_name" {
                blocks[i].Fields[j].Value = PREFIX + blocks[i].Fields[j].Value
            }
        }
    }
}


func parseFile(path string) ([]JobBlock, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var blocks []JobBlock
    var current JobBlock

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())

        if line == "" {
            if len(current.Fields) > 0 {
                blocks = append(blocks, current)
                current = JobBlock{}
            }
            continue
        }

        parts := strings.SplitN(line, ":", 2)
        if len(parts) == 2 {
            key := strings.TrimSpace(parts[0])
            value := strings.TrimSpace(parts[1])
            current.Fields = append(current.Fields, Field{Key: key, Value: value})
        }
    }

    if len(current.Fields) > 0 {
        blocks = append(blocks, current)
    }

    return blocks, scanner.Err()
}

func writeFile(path string, blocks []JobBlock) error {
    file, err := os.Create(path)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := bufio.NewWriter(file)
    for _, block := range blocks {
        for _, field := range block.Fields {
            fmt.Fprintf(writer, "%s: %s\n", field.Key, field.Value)
        }
        fmt.Fprintln(writer)
    }
    return writer.Flush()
}



func go() {
    blocks, err := parseFile("jobs.txt")
    if err != nil {
        panic(err)
    }

    modifyBlocks(blocks)

    err = writeFile("jobs_modified.txt", blocks)
    if err != nil {
        panic(err)
    }

    fmt.Println("Fichier modifié avec succès.")
}
