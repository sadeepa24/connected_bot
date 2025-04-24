package botapi

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestMsgCommonRead(t *testing.T) {
	btns := NewButtons([]int16{2})
	btns.AddBtcommon("Hello")
	btns.AddBtcommon("World")
	
	rd := &bytes.Buffer{}
	
	

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
			Message_thread_id: 2,
			Parse_mode: "Markdown",
			Reply_markup: btns.Getkeyboard(),
			Text: "Hello, wor2ld! Hello, worlfdrvrfderfrsewdd!Hello, <><><>< wsorld!Hello, world!Hedcfedsrello, world!Hello, wssorld!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!Hello, world!",	
		}
		rd.Reset()
		rd.Write([]byte("Hello This Text From Reader"))
		r.inite = false
		r.rover = false
		st := time.Now()
		r.Read(bt)
		if time.Since(st).Nanoseconds() > 0 {
			all += time.Since(st).Nanoseconds()
			fmt.Println(time.Since(st).Nanoseconds())
			ct++
		}
	}

	fmt.Println()
	fmt.Println(ct, (time.Duration(all) * time.Nanosecond).Milliseconds())

	// n, err = r.Read(bt)
	// fmt.Println(string(bt[:n]), err)
	// fmt.Println()
	// n, err = r.Read(bt)
	// fmt.Println(string(bt[:n]), err)
	// fmt.Println(time.Since(st).Nanoseconds())
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