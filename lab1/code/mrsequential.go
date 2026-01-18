package main

// 这个是串行版本的 MapReduce 框架，原本提供的

//
// simple sequential MapReduce.
//
// go run mrsequential.go wc.so pg*.txt
//

import "fmt"
import "6.5840/mr"
import "plugin"
import "os"
import "log"
import "io/ioutil"
import "sort"

// for sorting by key.
// 这里的意思是定义一个类型 ByKey，它是 mr.KeyValue 的切片。
// mr.KeyValue 是一个结构体，包含两个字段：Key 和 Value，分别表示键和值。
type ByKey []mr.KeyValue

// for sorting by key.
// 这里是 go 的 sort.Interface 接口的三个方法的实现
// 意思是对 ByKey 类型的切片进行排序时，会根据 Key 字段从小到大开始排序。
// 也就是统计完毕之后把所有一样的 Key 放在一起，方便后续的 Reduce 操作。
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

func main() {
	// 检查命令行参数是否足够
	// 如果参数不足，打印用法信息并退出程序
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: mrsequential xxx.so inputfiles...\n")
		os.Exit(1)
	}

	//
	// 加载插件，获取 Map 和 Reduce 函数，这个就是提取
	// 提取插件中的 Map 和 Reduce 函数
	//
	mapf, reducef := loadPlugin(os.Args[1])

	//
	// read each input file,
	// pass it to Map,
	// accumulate the intermediate Map output.
	//
	// Map 操作，首先初始化一个空的切片 intermediate 用于存储中间结果
	// 然后遍历每个输入文件，打开文件并读取内容
	//
	intermediate := []mr.KeyValue{}
	for _, filename := range os.Args[2:] {
		file, err := os.Open(filename)
		if err != nil {
			log.Fatalf("cannot open %v", filename)
		}
		content, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatalf("cannot read %v", filename)
		}
		file.Close()
		// Map 函数被调用，传入文件名和文件内容
		// Map 函数返回一个 mr.KeyValue 切片
		// 然后将这些中间结果追加到 intermediate 切片中
		// 这样，所有输入文件的中间结果都被收集到了 intermediate 切片中
		// 以便后续的 Reduce 操作使用
		kva := mapf(filename, string(content))
		intermediate = append(intermediate, kva...)
	}

	//
	// a big difference from real MapReduce is that all the
	// intermediate data is in one place, intermediate[],
	// rather than being partitioned into NxM buckets.
	//
	// 这里是直接 append，但是在真正的分布式操作系统里会分成很多份
	//

	// 把 ACABD 按照 Key 排序
	// AABCD
	sort.Sort(ByKey(intermediate))

	oname := "mr-out-0"
	ofile, _ := os.Create(oname)

	//
	// call Reduce on each distinct key in intermediate[],
	// and print the result to mr-out-0.
	//
	// Reduce 操作
	// 其实这里就是双指针操作，从 i 的下一个位置 j 开始往后找
	// 前提就是排序过的，所以相同 Key 的元素都是挨在一起的
	//
	i := 0
	for i < len(intermediate) {
		j := i + 1
		// 只要 j 没越界，且 j 位置的 Key 和 i 位置的 Key 相同，j 就一直往后移		
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		// 此时，从 i 到 j-1 的所有元素，它们的 Key 都是一样的！
		// 这一段范围 [i, j) 就是我们要处理的同一个 Key 的所有数据。
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		// 把这个 Key 和它所有的 values 交给 Reduce 函数
		output := reducef(intermediate[i].Key, values)

		// this is the correct format for each line of Reduce output.
		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)
		// 处理完这一批相同的 Key 后，i 直接跳到 j 的位置，开始处理下一个不同的 Key
		i = j
	}

	ofile.Close()
}

// load the application Map and Reduce functions
// from a plugin file, e.g. ../mrapps/wc.so
// 这里用到了 Go 语言的 plugin 包。它允许在运行时动态加载编译好的代码。
// 流程：打开文件 -> 找名字叫 "Map" 的东西 ->
// 确认它真的是个函数 -> 返回给主程序使用。
func loadPlugin(filename string) (func(string, string) []mr.KeyValue, func(string, []string) string) {
	p, err := plugin.Open(filename)
	if err != nil {
		log.Fatalf("cannot load plugin %v", filename)
	}
	xmapf, err := p.Lookup("Map")
	if err != nil {
		log.Fatalf("cannot find Map in %v", filename)
	}
	mapf := xmapf.(func(string, string) []mr.KeyValue)
	xreducef, err := p.Lookup("Reduce")
	if err != nil {
		log.Fatalf("cannot find Reduce in %v", filename)
	}
	reducef := xreducef.(func(string, []string) string)

	return mapf, reducef
}
