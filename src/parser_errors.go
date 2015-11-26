package main

/*
 *  Add to .gitignore
 *     *.gen.go
 */

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const errorList_sfx = "ErrorList"
const replyList_sfx = "ReplyList"
const errgen_sfx = "_err.gen.go"

var skip_paths = []string{"golang.org", "gopkg.in", "github.com", "google.com"}

func GenErrorInSrc(src_path string) {


  log.Println(src_path)
	//base_dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	//base_dir = base_dir + src_path
	//src_path, _ = filepath.Abs(base_dir)

	filepath.Walk(src_path, func(path string, info os.FileInfo, err error) error {
		// смотрим все файлы *.go, но не _gen.go
		if strings.HasSuffix(path, errgen_sfx) {
			println("remove: ", path)
			os.Remove(path)
		}
		return nil
	})

	filepath.Walk(src_path, func(path string, info os.FileInfo, err error) error {

		// смотрим все файлы *.go, но не _err.gen.go
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, errgen_sfx) {
			return nil
		}

		// пропускаем папки
		for _, skip_path := range skip_paths {
			if strings.Contains(path, skip_path) {
				return nil
			}
		}

		parse_file(path)
		return nil
	})
}

type Field struct {
	Name    string
	Line    int
	Comment string
	Type    string
}

type StructData struct {
	Name   string
	Fields []Field
}

type GoFile struct {
	Name       string
	Path       string
	SourcePath string
	Package    string
	Src        string

	err_structs   []StructData
	reply_structs []StructData
}

var codeError = 7000

var regExpFindCode = regexp.MustCompile(`[^\S\n\r]*CODE:\d+[^\S\n\r]*`)

// находим все первая CODE:123 строка
var regExpFindFirstCode = regexp.MustCompile(`^[\t|\S]*[^(CODE:\d+)]+(CODE:\d+)`)

// находим все не первые CODE:123 строки
var regExpFindOthersCode = regexp.MustCompile(`^[\t|\S]*[^(CODE:\d+)]+CODE:\d+ | (CODE:\d+)`)

// первый коммент // или /* от начала строки
var regExpFindComment = regexp.MustCompile(`^[^\/]+(\/\/)|^[^\/]+(\/\*)+`)

func saveStruct(goFile *GoFile, insert_in_source bool) error {

	println(goFile.Path)

	var lines []string

	if insert_in_source {
		input, err := ioutil.ReadFile(goFile.SourcePath)
		if err != nil {
			log.Fatalln(err)
		}
		lines = strings.Split(string(input), "\n")
	}

	file, err := os.Create(goFile.Path)
	if err != nil {
		println(err.Error())
		return err
	}
	file.WriteString("package " + goFile.Package + "\n")
	file.WriteString("\n")
	if goFile.Package != "types" {
		file.WriteString("import (\n")

		add_rpc_types := false
		for _, struct_data := range goFile.err_structs {
			for _, field := range struct_data.Fields {
				if field.Type == "RPCError" {
					add_rpc_types = true
				}
			}
		}
		if add_rpc_types {
			file.WriteString("    . \"se/rpc/types\"\n")
		}

		add_errors := false
		for _, struct_data := range goFile.err_structs {
			for _, field := range struct_data.Fields {
				if field.Type == "error" {
					add_errors = true
				}
			}
		}
		if add_errors {
			file.WriteString("    \"errors\"\n")
		}
		file.WriteString(")\n")
	}

	file.WriteString("\n")

	for _, struct_data := range goFile.err_structs {

		list_name := strings.TrimSuffix(struct_data.Name, "List")
		file.WriteString("var " + list_name + " = " + struct_data.Name + "{\n")
		for _, field := range struct_data.Fields {

			codeError++
			code_num_srt := strconv.Itoa(codeError)
			code_str := "CODE:" + code_num_srt

			if field.Type == "RPCError" {
				if insert_in_source {
					line_num := field.Line - 1
					line := lines[line_num]
					line = strings.Replace(line, "\n", "", -1)
					line = strings.Replace(line, "\r", "", -1)
					line = strings.Replace(line, "\r\n", "", -1)
					line = strings.Replace(line, "\n\r", "", -1)

					//log.Println("origin: ", line)
					line = regExpFindCode.ReplaceAllString(line, " ")
					{
						regexp_str := regExpFindComment.FindStringSubmatch(line)
						if len(regexp_str) > 1 { // коммент есть
							comment_start := regexp_str[0]
							replace_str := comment_start + " " + code_str
							if strings.HasSuffix(comment_start, "/*") {
								comment_start = strings.TrimSuffix(comment_start, "/*")
								replace_str = comment_start + "// " + code_str + " /*"
							}
							line = regExpFindComment.ReplaceAllString(line, replace_str)
						} else { // коммента нет
							line += " // " + code_str
						}
					}
					// log.Println("         ", line)
					lines[line_num] = line
				}
			}

			clear_comment := regExpFindCode.ReplaceAllString(field.Comment, "")
			clear_comment = strings.Replace(clear_comment, "\n", "", -1)
			clear_comment = strings.Replace(clear_comment, "\r", "", -1)
			clear_comment = strings.Replace(clear_comment, "\r\n", "", -1)
			clear_comment = strings.Replace(clear_comment, "\n\r", "", -1)

			str := "    " + field.Name + ": "
			if field.Type == "error" {
				str += "errors.New(\"" + field.Name + "\"),\n"
			}

			if field.Type == "RPCError" {
				str += "RPCError{\n"
				str += "        Code:" + code_num_srt + ",\n"
				str += "        Id:\"" + field.Name + "\",\n"
				str += "        Description:\"" + clear_comment + "\",\n"
				str += "        Class:\"" + list_name + "\",\n"
				str += "    },\n"
			}

			file.WriteString(str)
		}
		file.WriteString("}\n")
	}

	for _, struct_data := range goFile.reply_structs {
		list_name := strings.TrimSuffix(struct_data.Name, "List")
		file.WriteString("var " + list_name + " = " + struct_data.Name + "{\n")
		for _, field := range struct_data.Fields {
			str := "    " + field.Name + ": \"" + field.Name + "\",\n"
			file.WriteString(str)
		}
		file.WriteString("}\n")
	}

	file.Close()

	if insert_in_source {
		output := strings.Join(lines, "\n")
		output = strings.Replace(output, "\r\n", "\n", -1)
		output = strings.Replace(output, "\n", "\r\n", -1)

		src_data, err := ioutil.ReadFile(goFile.SourcePath)
		if err != nil {
			log.Fatalln(err)
		}
		src_str := string(src_data)

		if output != src_str {
			log.Println("Output: ", goFile.SourcePath)
			err = ioutil.WriteFile(goFile.SourcePath, []byte(output), 0644)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
	return nil
}

func typeName(arg interface{}, struct_data *StructData, f *token.FileSet, src string) string {
	switch v := arg.(type) {
	case *ast.ChanType:
		return "chan " + typeName(v.Value, struct_data, f, src)
	case *ast.ArrayType:
		res := "array["
		if v.Len != nil {
			res += typeName(v.Len, struct_data, f, src)
		}
		res += "]" + typeName(v.Elt, struct_data, f, src)
		return res
	case *ast.FuncType:
		res := "func("
		if v.Params != nil {
			lenParam := len(v.Params.List)
			for i, param := range v.Params.List {
				if len(param.Names) > 0 {
					res += param.Names[0].Name + " "
				}
				res += typeName(param.Type, struct_data, f, src)
				if i+1 != lenParam {
					res += ", "
				}
			}
		}
		res += ")("
		if v.Results != nil {
			lenResult := len(v.Results.List)
			for i, result := range v.Results.List {
				if len(result.Names) > 0 {
					res += result.Names[0].Name + " "
				}
				res += typeName(result.Type, struct_data, f, src)
				if i+1 != lenResult {
					res += ", "
				}
			}
		}
		res += ")"
		return res
	case *ast.Ident:
		return v.Name
	case *ast.InterfaceType:
		res := "interface{\n"
		if v.Methods != nil {
			for _, method := range v.Methods.List {
				if len(method.Names) > 0 {
					res += "  " + method.Names[0].Name
				}
				res += "  " + typeName(method.Type, struct_data, f, src)
				res += "\n"
			}
		}
		res += "}"
		return res
	case *ast.MapType:
		return "map[" + typeName(v.Key, struct_data, f, src) + "]" + typeName(v.Value, struct_data, f, src)
	case *ast.StructType:
		res := "struct{\n"
		if v.Fields != nil {
			for _, method := range v.Fields.List {
				if len(method.Names) > 0 {
					res += "  " + method.Names[0].Name
					var field Field
					field.Name = method.Names[0].Name
					field.Line = f.Position(method.Pos()).Line
					field.Comment = method.Comment.Text()
					field.Type = src[method.Type.Pos()-1 : method.Type.End()-1]
					struct_data.Fields = append(struct_data.Fields, field)
				}
				res += "  " + typeName(method.Type, struct_data, f, src)
				if method.Tag != nil {
					res += "  " + method.Tag.Value
				}
				res += "\n"
			}
		}
		res += "}"
		return res
	case *ast.StarExpr:
		return "*" + typeName(v.X, struct_data, f, src)
	default:
		return reflect.TypeOf(arg).String()
	}
	return ""
}

func parse_file(file_path string) {

	buf := bytes.NewBuffer(nil)
	file, err := os.Open(file_path)
	if err != nil {
		return
	}
	io.Copy(buf, file)
	file.Close()

	src := string(buf.Bytes())

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return
	}

	var goFile GoFile
	goFile.Src = src
	goFile.Package = f.Name.Name
	goFile.SourcePath = file_path
	goFile.Path = file_path[:len(file_path)-3] + errgen_sfx

	for _, decl := range f.Decls {
		if tmp, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range tmp.Specs {
				if tmp, ok := spec.(*ast.TypeSpec); ok {
					if structType, ok := tmp.Type.(*ast.StructType); ok {

						name := tmp.Name.Name

						var struct_data StructData
						struct_data.Name = name
						if strings.HasSuffix(name, errorList_sfx) {
							typeName(structType, &struct_data, fset, src)
							goFile.err_structs = append(goFile.err_structs, struct_data)
						}
						if strings.HasSuffix(name, replyList_sfx) {
							typeName(structType, &struct_data, fset, src)
							goFile.reply_structs = append(goFile.reply_structs, struct_data)
						}

					}
				}
			}
		}
	}

	if len(goFile.err_structs) > 0 {
		saveStruct(&goFile, false)
	}
}
