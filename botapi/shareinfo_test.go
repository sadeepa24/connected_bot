package botapi

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestMsgCommonRead(t *testing.T) {
	btns := NewButtons([]int16{2})
	btns.AddBtcommon("Hello")
	btns.AddBtcommon("World")
	
	rd := &bytes.Buffer{}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func ()  {
		bt := make([]byte, 500)
		all := int64(0)
		ct := 0 
		for i := 0; i < 100000; i++ {
			r := &Msgcommon{
				reader: rd,
				rover: false,
				inite: false,
				Infocontext: &Infocontext{
					ChatId: 123456789,
					User_id: 987654321,
				},
				Meadiacommon: &Meadiacommon{
					Photo: "photsoid",
					Video: "videoid",
					Caption: "this is caption",
					Media: &InputMedia{
						Type: "media typsse",
						Media: "real media",
						//Caption: "inside caption",
					},
				},
				Message_id: 3,
				Message_thread_id: 56,
				Parse_mode: "Markdown",
				Reply_markup: btns.Getkeyboard(),
				Text: "Hello, wor2ld! Hdello, worlfdrvrfderfrseswdd!Hello, <><><>< wsorld!Hello, world!Hedcfedsrello, world!Hello, wssorld!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!",	
			}
			rd.Reset()
			rd.Write([]byte("Hello This Text From Reader"))
			r.inite = false
			r.rover = false
			st := time.Now()
			r.Read(bt)
			if time.Since(st).Nanoseconds() > 0 {
				all += time.Since(st).Nanoseconds()
				//fmt.Println(time.Since(st).Nanoseconds())
				ct++
		}
		
		}
		wg.Done()
		fmt.Println()
		fmt.Println(ct, (time.Duration(all) * time.Nanosecond).Milliseconds())
	}()
	go func ()  {
		bt := make([]byte, 500)
		all := int64(0)
		ct := 0 
		for i := 0; i < 100000; i++ {
			r := &Msgcommon{
				reader: rd,
				rover: false,
				inite: false,
				Infocontext: &Infocontext{
					ChatId: 123456789,
					User_id: 987654321,
				},
				Meadiacommon: &Meadiacommon{
					Photo: "photsoid",
					Video: "videoid",
					Caption: "this is caption",
					Media: &InputMedia{
						Type: "media typsse",
						Media: "real media",
						//Caption: "inside caption",
					},
				},
				Message_id: 3,
				Message_thread_id: 56,
				Parse_mode: "Markdown",
				Reply_markup: btns.Getkeyboard(),
				Text: "Hello, wor2ld! Hdello, worlfdrvrfderfrseswdd!Hello, <><><>< wsorld!Hello, world!Hedcfedsrello, world!Hello, wssorld!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!",	
			}
			rd.Reset()
			rd.Write([]byte("Hello This Text From Reader"))
			r.inite = false
			r.rover = false
			st := time.Now()
			r.Read(bt)
			if time.Since(st).Nanoseconds() > 0 {
				all += time.Since(st).Nanoseconds()
				//fmt.Println(time.Since(st).Nanoseconds())
				ct++
		}
		
		}
		wg.Done()
		fmt.Println()
		fmt.Println(ct, (time.Duration(all) * time.Nanosecond).Milliseconds())
	}()
	go func ()  {
		bt := make([]byte, 500)
		all := int64(0)
		ct := 0 
		for i := 0; i < 100000; i++ {
			r := &Msgcommon{
				reader: rd,
				rover: false,
				inite: false,
				Infocontext: &Infocontext{
					ChatId: 123456789,
					User_id: 987654321,
				},
				Meadiacommon: &Meadiacommon{
					Photo: "photsoid",
					Video: "videoid",
					Caption: "this is caption",
					Media: &InputMedia{
						Type: "media typsse",
						Media: "real media",
						//Caption: "inside caption",
					},
				},
				Message_id: 3,
				Message_thread_id: 56,
				Parse_mode: "Markdown",
				Reply_markup: btns.Getkeyboard(),
				Text: "Hello, wor2ld! Hdello, worlfdrvrfderfrseswdd!Hello, <><><>< wsorld!Hello, world!Hedcfedsrello, world!Hello, wssorld!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!",	
			}
			rd.Reset()
			rd.Write([]byte("Hello This Text From Reader"))
			r.inite = false
			r.rover = false
			st := time.Now()
			r.Read(bt)
			if time.Since(st).Nanoseconds() > 0 {
				all += time.Since(st).Nanoseconds()
				//fmt.Println(time.Since(st).Nanoseconds())
				ct++
		}
		
		}
		wg.Done()
		fmt.Println()
		fmt.Println(ct, (time.Duration(all) * time.Nanosecond).Milliseconds())
	}()
	wg.Wait()
	// n, err = r.Read(bt)
	// fmt.Println(string(bt[:n]), err)
	// fmt.Println()
	// n, err = r.Read(bt)
	// fmt.Println(string(bt[:n]), err)
	// fmt.Println(time.Since(st).Nanoseconds())

	//without make
	// 532 268
	// 510 282
	// 547 295
	
	//with make
	//422 417
	//434 345
	//444 335

	//with pool
	//708 280
	//610 254
	//624 294

}


// func TestPrebuf(t *testing.T) {
// 	prebuf := PreBuuf{
// 		Buffer: &bytes.Buffer{},
// 	}

// 	prebuf.Write(
		

	
// 	r := make([]byte, 4000)
// 	prebuf.Read(r)
// 	fmt.Println(string(r[3950:]))
// 	prebuf.Read(r)
// 	fmt.Println()
// 	fmt.Println(string(r[0:100]))
// }