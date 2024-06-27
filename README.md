# brief about #com_municat

Sẽ ra sao nếu như một người, chỉ mới chân ướt chân ráo biết tí về Go, lại phải làm một dự án đủ khó như một phần mềm gọi điện, dùng toàn bộ sức lực, thời gian để tìm, research và viết một chương trình đủ lớn?

Và đó là cách câu chuyện bắt đầu, từ #com_municat.

Đây chỉ là tóm lược dựa trên bản mới nhất được push lên GitHub gần đây, chưa bao gồm quá trình làm ra phần mềm. 

references…, thanks for helping me through this funny experiment…

[https://github.com/Bronstein0x113c1c3/2016](https://github.com/Bronstein0x113c1c3/2016)

[https://github.com/kechako/vchat](https://github.com/kechako/vchat)

[https://github.com/joan41868/VoGRPC](https://github.com/joan41868/VoGRPC)

# i. diagram

![Untitled Diagram.drawio(3).svg](diagram/Untitled_Diagram.drawio(3).svg)

Mô hình vận hành của ý tưởng trên sẽ giống như thế này, phía màu tím là những thành phần thuộc về client và màu cam thuộc về server, đường đi của gói tin được đánh dấu bằng 2 màu xanh và trắng, trong đó màu xanh là hướng gửi và màu trắng là hướng nhận. Đây là mô hình đơn giản nhất cho toàn bộ quá trình truyền tải, chưa bao gồm các quá trình điều khiển từ phía client và quản lý từ phía server…..

Để giải thích về sự xuất hiện của các buffer như send buffer hay recv buffer hay buffer channel thì các buffer này, có bản chất là các buffered channel ở trong Go, loại buffer mà cho khả năng cho phép tạm thời các data được đọng lại trong channel chừng nào channel còn sức chứa, cũng phần nào giống cơ chế message queue. Việc sử dụng buffer có một vài tác dụng tích cực đến hiệu suất tổng thể giữa server và client, bao gồm:

- **Hoạt động tốt hơn trong các điều kiện mạng không tốt:** do các channel có nhiệm vụ lưu trữ trước các dữ liệu, đặc biệt trong trường hợp sử dụng HTTP/3, vốn sử dụng UDP để truyền tải, như vậy sẽ tránh được việc bị mất gói tin liên tục.
- **Tăng cường kiểm soát đến các thành phần nhỏ hơn trong cả server và client**: khi thêm phần buffer, việc xử lý nhập xuất (I/O) được tách ra, không can hệ trực tiếp đến phần truyền tải (Client), từ đó dễ dàng kiểm soát các cơ chế này dễ dàng hơn mà không ảnh hưởng trực tiếp lẫn nhau.
- **Giảm thiểu áp lực xử lý lên server và client cũng như chống lại starvation**: Xử lý đồng thời (concurrency) là một bài toán đau đầu và ức chế cho những người làm việc với Go, đặc biệt là các vấn đề sử dụng luồng dữ liệu cũng như tối ưu. Trong trường hợp này, nếu như không có sự xuất hiện của các buffered channel thì có lẽ, server sẽ liên tục bị dồn do lượng gói tin gửi liên tục cực lớn và server sẽ không thể xử lý kịp thời rồi gửi cho các client, dẫn đến hiệu năng bị chậm đi. Nhìn từ góc độ client, việc lưu trữ buffer này giúp cho client không bị nghẽn nếu như có quá nhiều gói tin được gửi về từ phía client. Một tác dụng lớn hơn nữa là các buffered channel này sẽ giúp server và client tránh việc bị lỗi khi quyết định đóng channel, giống như trên, việc data gửi liên tục tới unbuffered channel có thể khiến phần mềm bị crash khi channel liên tục bị đọng lại data và cả trường hợp data cố tình gửi vào closed channel.

Dù vậy, với các server và client xử lý đủ nhanh thì có vẻ các biệt hiệu năng sẽ không rõ ràng lắm và trên lý thuyết, có thể việc sử dụng buffered channel sẽ chậm hơn so với unbuffered channel, tùy một số trường hợp nhất định. Suy cho cùng, để tối ưu làm sao cho hiệu quả và gây ra ít lỗi nhất thì việc lưu trữ này cũng là một ý không hề tệ và nhờ vậy, các thành phần hoạt động không bị chồng chéo lên nhau.

Đó là mô hình đơn giản nhất cho phần mềm trò chuyện này, giờ ta sẽ sang các thành phần khác của project, bao gồm client, server và protobuf.

# ii. protobuf

### the protobuf

```protobuf
syntax = "proto3";
option go_package="/protobuf";
//one server, many clients....

message ClientMSG {
    bytes chunk = 1;
    string name = 2;
}
message ClientSignal{}
message ClientREQ{
    oneof request{
        ClientMSG message =1;
        ClientSignal signal =2;
    }
}
message ServerMSG{
    ClientMSG msg = 1;
    string id = 2;    
}
message ServerSignal{}
message ServerRES{
    oneof response{
        ServerMSG message =1;
        ServerSignal signal =2;
    }
}

service Calling {
    rpc VoIP (stream ClientREQ) returns (stream ServerRES);
}
```

Vì phần protobuf khá dài nên tôi sẽ cố gắng chia thành 2 phần, server và client….

## client protobuf

```protobuf
message ClientMSG {
    bytes chunk = 1;
    string name = 2;
}
message ClientSignal{}
message ClientREQ{
    oneof request{
        ClientMSG message =1;
        ClientSignal signal =2;
    }
}
```

Như ta có thể thấy, một request của client có 2 loại, là Signal và MSG, trong đó, Signal có vai trò để thử tín hiệu truy cập ngay khi thiết lập kết nối. Việc sửa đổi như này xuất phát từ việc thêm tính năng interceptor giới hạn số lượng truy cập và xác thực. Trong trường hợp này, cái tín hiệu thử sẽ không chứa thông tin gì, nhưng có nhiệm vụ thông báo cho server về việc kết nối đã thành công hay chưa, nếu mà server không trả về thông tin ngay lập tức thì tức là đã gặp sự cố. Sau khi chương trình thành công, các request này sẽ được gửi đi các đoạn MSG bao gồm giọng nói và tên của người dùng. Ban đầu, tôi có ý định tách riêng và không gộp chung như trên, nhưng việc test kết nối ban đầu là việc phải làm cũng như giảm tải cho server thì việc gộp vào với oneof là hợp lý nhất. Nhờ đó, tôi gần như giảm thiểu trường hợp bị lỗi khi client thiết lập IO mà không thể kết nối tới server. 

## server

```protobuf
message ServerMSG{
    ClientMSG msg = 1;
    string id = 2;    
}
message ServerSignal{}
message ServerRES{
    oneof response{
        ServerMSG message =1;
        ServerSignal signal =2;
    }
}
service Calling {
    rpc VoIP (stream ClientREQ) returns (stream ServerRES);
}
```

# iii. implementation

Đến phần triển khai, nhiệm vụ của 2 phía cũng khác nhau rõ rệt, điểm rõ nét nhất từ cái diagram trên là một server phải đảm nhận nhiều connections từ client, cũng như việc client kết nối tới client và nhận từ phía server. Thêm nữa, ngoài các công đoạn xử lý còn có các cơ chế ngắt phần mềm, cả từ phía người dùng, tự động ngắt, cơ chế kiểm tra kết nối,…. Đây sẽ là phần cuối cùng của bản mô tả.

## 1. client implementation.

Trọng tâm của bản com_municat này so với dự án the_call đầu tiên là phần client, thay vì tôi cho 1 input và hàng đống output stream gây nặng máy và thiếu đồng bộ thì giờ tôi chỉ cần 1 input và 1 output, nhờ vậy, chương trình chạy mượt hơn kha khá, thêm nữa, việc tách quá trình gửi và nhận thành 4 quá trình nhỏ hơn, và chỉ liên kết với nhau thông qua buffer cũng giúp chương trình dễ triển khai và cho hiệu suất tốt hơn trước. Ngoài ra, cơ chế kiểm soát cũng được thêm vào để giúp cho phần mềm hoạt động hiệu quả hơn như một hệ quả của việc phân chia tác vụ rõ ràng hơn….

### 1.1. input - sending

Phần input - sending này được chia thành 2 phần riêng biệt và hoạt động độc lập, là input và sending. 

Input có nhiệm vụ là thu giọng nói, encode, serialize và buffering, dưới đây là các phần của input cũng như sending….

### constructor

```go
func InputInit(channel int, sample_rate float32, buffer_size int, data_length int, wg *sync.WaitGroup) (*Input, error) {
	portaudio.Initialize()
	buf := make([]int16, buffer_size)
	streamer, err := portaudio.OpenDefaultStream(channel, 0, float64(sample_rate), buffer_size, &buf)
	if err != nil {
		portaudio.Terminate()
		return nil, err
	}
	encoder, err := opus.NewEncoder(int(sample_rate), channel, opus.AppVoIP)
	if err != nil {
		portaudio.Terminate()
		return nil, err
	}
	// len := int(int(sample_rate) / channel / 1000/)
	data_chan := make(chan []byte, data_length)
	return &Input{
		stream:    streamer,
		encoder:   encoder,
		wg:        wg,
		buf:       buf,
		signal:    make(chan struct{}),
		data_chan: data_chan,
		byte_len:  data_length,
	}, nil
}
```

Đây là phần constructor của input, việc đầu tiên là khởi tạo buffer âm thanh định dạng int16 signed, với buffer_size theo tham số được truyền vào, cùng với số kênh, tốc độ mẫu. Điểm mới nằm ở việc xuất hiện encoder Opus để nén lại âm thanh phù hợp với mức âm thanh thấp, loại bỏ tiếng vang cũng như các tạp âm, có thể nói rằng sự xuất hiện của encoder này đã thay đổi số phận của dự án theo hướng tích cực và cũng như dễ triển khai sau khi tách toàn bộ thành phần. Cuối cùng, các dữ liệu được buffer lại bằng data_chan có capacity là 1000 để giúp cho quá trình xử lý âm thanh liên tục không gây ra các hiện tượng tiêu cực đến tốc độ xử lý. Cuối cùng, là thêm wg vào bộ input để kiểm soát cơ chế concurrency hiệu quả hơn và gom hết các thông tin kia lại với nhau, thành một bộ input hoàn chỉnh.

### process

```go
func (i *Input) Process() {
	i.stream.Start()
	defer i.wg.Done()
	// defer i.close(i.data_chan)
	defer close(i.data_chan)
	defer portaudio.Terminate()
	// defer i.stream.Stop()
	for {
		i.stream.Read()
		data := make([]byte, i.byte_len)
		n, err := i.encoder.Encode(i.buf, data)
		if err != nil {
			return
		}
		select {
		case <-i.signal:
			return
		case i.data_chan <- data[:n]:
		}
	}
}
func (i *Input) GetChannel() chan []byte {
	return i.data_chan
}
```

Bắt đầu với quá trình xử lý input, thì ta bắt đầu nhận giọng nói với i.stream.Read(), 3 hàm defer với mục đích dọn dẹp lại input sau khi quá trình xử lý này kết thúc. Mọi chuyện bắt đầu thực sự với hàm for, với hàm i.stream.Read(), input xử lý giọng nói rồi ghi vào buffer, data cũng được thêm mới mỗi lần lặp để tránh quá trình gán đi gán lại, vốn mất nhiều thời gian, hàm encoding để xử lý âm thanh gốc thành dạng âm thanh được nén vào data và cuối cùng, quá trình gửi sang buffer cũng như ngắt input được thực hiện bằng lệnh select. Tất cả đều phải được xử lý bằng một goroutine riêng biệt để quá trình này hoạt động song song với output và các thành phần khác. Và dĩ nhiên, không thể quên việc đưa ra channel để phần sending có thể tận dụng rồi gửi cho server gói tin.

Từ việc tách ra thành phần, việc viết input trở nên dễ dàng đến lạ thường….

### sending

```go
func Send(data_chan chan []byte, client pb.Calling_VoIPClient, name string) {
	// data_chan := input.GetChannel()
	for data := range data_chan {
		// client.Send(&pb.ClientMSG{
		// 	Chunk: data,
		// 	Name:  name,
		// })
		client.Send(&pb.ClientREQ{
			Request: &pb.ClientREQ_Message{
				Message: &pb.ClientMSG{
					Chunk: data,
					Name:  name,
				},
			},
		})
	}
}
```

Như một hệ quả, phần sending có nhiệm vụ gửi các gói tin lại cho server từ data_chan, việc sử dụng range trong Go có tác dụng rất lớn khi client chỉ việc gửi đi những gì có trong data_chan và nếu data_chan có ngắt thì hàm for này cũng dừng theo. Hàm viết khá đơn giản, chỉ là bao gồm các thông tin đủ cần thiết rồi gửi đi thôi… Cũng như input, hàm send cũng phải hoạt động trong một luồng riêng biệt để hoạt động độc lập và nhanh nhất có thể.

### 1.2. receiving - output

Tương tự input - sending, thì receiving - output có nhiệm vụ lấy gói tin về, giải mã chúng rồi phát ra âm thanh. Các nhiệm vụ bao gồm receiving, buffering, deserialize, decode và phát ra âm thanh. Xem ra, receiving - output làm đơn giản hơn là vì việc chỉ lấy âm thanh rồi phát ra tiếng cũng không tốn công triển khai, chưa kể cấu trúc code cũng tương tự như phần trước.

### constructor

```go
func OutputInit(channel int, sample_rate float32, buffer_size int, data_chan chan []byte) (*Output, error) {
	buf := make([]int16, buffer_size)
	portaudio.Initialize()
	streamer, err := portaudio.OpenDefaultStream(0, channel, float64(sample_rate), buffer_size, &buf)
	if err != nil {
		portaudio.Terminate()
		return nil, err
	}
	decoder, err := opus.NewDecoder(int(sample_rate), channel)
	if err != nil {
		portaudio.Terminate()
		return nil, err
	}
	return &Output{
		buf:       buf,
		stream:    streamer,
		decoder:   decoder,
		data_chan: data_chan,
	}, nil
}
```

Cũng giống như phần input - sending, nhưng giờ cơ chế được đổi khác. Đáng ra, tôi có thể ghép cái stream này vào làm một, nhưng tôi sợ rằng nó xảy ra các xung đột nên quyết định tạo ra một stream khác chỉ để phát ra loa, như việc input channel là 0. Các tham số vẫn như vậy, chỉ khác là, giờ decoder sẽ xuất hiện, lấy các tham số như tốc độ mẫu với channel. Cuối cùng là lấy đủ thông tin, vậy thôi.

### receiving

```go
func Receive(data_chan chan []byte, ctx context.Context, client pb.Calling_VoIPClient, stop context.CancelFunc) {

	for {
		select {
		case <-ctx.Done():
			return
		default:
			segment, err := client.Recv()
			if segment.GetSignal() != nil {
				continue
			}
			data := segment.GetMessage()
			if err != nil {
				stop()
				return
			}
			data_chan <- data.Msg.Chunk
		}
	}
}

```

Phần receiving thì có vẻ nhẹ nhàng hơn, việc sử dụng context giúp cho chương trình có cơ chế ngắt tốt hơn rất nhiều và client giờ chỉ tập trung nhận gói tin, gửi chúng vào data_chan và chờ cho output nhận được rồi phát ra âm thanh, tôi có sửa lại chút thì cho thêm cả quá trình check lỗi và ngắt chương trình ngay khi không thể kết nối tới server, sử dụng chung tín hiệu ngắt. Nói chung là…. dễ vl…. 

### process

```go
func (o *Output) Process() {
	// defer o.wg.Done()
	o.stream.Start()
	for data := range o.data_chan {
		o.decoder.Decode(data, o.buf)
		o.stream.Write()
	}
}
```

Phần này thì hết nói, vì nó quá đơn giản :))))

### 1.3. control mechanism

Đến thời điểm tôi viết dự án này, tôi nhận thấy cơ chế kiểm soát, ở đây bao gồm cơ chế ngắt và cơ chế điều khiển vô cùng quan trọng khi chúng biến một chương trình khó hiểu, chỉ thuần truyền tải trở nên bình thường hơn rất nhiều…. Phần control có 2 phần phụ, là command - gửi mệnh lệnh điều khiển và control - cơ chế kiểm soát.

### command

```go
func Command(signal chan struct{}, command_chan chan string) {
	defer close(command_chan)
	for {
		select {
		case <-signal:
			return
		default:
			for {
				var command string
				fmt.Print("what command: ")
				fmt.Scanln(&command)
				command_chan <- command
			}
		}
	}
}
```

Cơ chế kiểm soát cũng đơn giản là vì việc của nó là chạy 1 vòng lặp, điền mệnh lệnh và chạy tiếp, cũng như nghe từ phía control, vốn sẽ nhận cả lệnh ngắt ctrl+c.

### control

```go
func Control(ctx context.Context, command chan string, signal chan struct{}, input *input.Input, output *output.Output, stop context.CancelFunc) {
	defer close(signal)
	for {
		select {
		case <-ctx.Done():
			log.Println("ctrl + c detected, exitting...")
			go input.Close()
			go output.Close()
			return
		case x, ok := <-command:
			if !ok {
				return
			}
			switch x {
			case "mute":
				log.Println("muting called")
				go input.Mute()
				continue
			case "unmute":
				log.Println("unmuting called")

				go input.Play()
				continue
			case "stop":
				log.Println("stopping called")
				go input.Close()
				go output.Close()
				stop()
				return
			default:

			}
			// var command
		}
	}
}
```

Quá trình điểu khiển lệnh ngắt này bao gồm việc nhận tín hiệu ngắt ctrl + c nếu muốn tắt khẩn cấp, và cơ chế nhận lệnh từ Command. Nhận dữ liệu từ command có 3 loại lệnh chính, là lệnh ngắt thu âm, thu âm và dừng, dẫu vậy, lệnh điều khiển này mới xong ở phần input, còn output thì vẫn là nghe nên sau này sẽ được bổ sung. Các lệnh này được tận dụng triệt để bằng cơ chế goroutine để đảm bảo là việc input không bị ảnh hưởng tiêu cực từ phân khu điều khiển, tăng tính độc lập cho các thành phần.

### what outside???

Dĩ nhiên, các tín hiệu này sẽ được làm từ trước, sau khi chương trình gọi được đến server. Tại đây, quá trình kiểm tra kết nối trước khi chuyển âm thanh được thêm vào nhằm mục đích ngắt chương trình ngay lập tức nếu như không kết nối thành công do nghẽn người hoặc là sai mật khẩu hoặc sự cố từ server… Tất cả các tín hiệu và quá trình này đều được thu gọn bằng một hàm Serve() có nhiệm vụ chạy tất cả các hàm với các goroutine hoạt động riêng rẽ, không can thiệp đến nhau.

```go
	{/**/
	log.Println("initiating the signal....")
	signal_chan, command_chan, ctx_signal, ctx_client, ctx1, ctx2, stop, cancel1, cancel2 := step.Signal(passcode)
	defer cancel1()
	defer cancel2()
	log.Println("signal initiation done, connecting to the server....")

	client, err := step.InitClient(host, ctx_client, "caller", http3)

	if err != nil {
		log.Fatalf("error before connecting: %v \n", err)
	}
	if err := step.TestConn(client); err != nil {
		log.Fatalf("error after connection: %v \n", err)
	}
	/*other things....*/
	wg.Add(1)
	step.Serve(input, output, data_chan, signal_chan, command_chan, client, ctx1, ctx2, name, stop)
	<-ctx_signal.Done()
	defer wg.Wait()
	/**/}
	
	
func Serve(input *input.Input, output *output.Output, data_chan chan []byte, signal_chan chan struct{}, command_chan chan string,
	client pb.Calling_VoIPClient, ctx1, ctx2 context.Context, name string, stop context.CancelFunc) {
	go input.Process()
	go transmitting.Send(input.GetChannel(), client, name)
	go transmitting.Receive(data_chan, ctx1, client)
	go output.Process()
	go control.Command(signal_chan, command_chan)
	go control.Control(ctx2, command_chan, signal_chan, input, output, stop)
}
```

### 1.4. simple authentication - metadata - context

Cùng với cơ chế ngắt nghỉ, sự xuất hiện của metadata với context là một thứ thay đổi cách chương trình vận hành, khi từ giờ, tôi có thể yêu cầu client điền mật khẩu và bắt đầu quá trình xác thực tiền kết nối, tiếp đến, nó cũng tăng khả năng phản hồi và khắc chế nhược điểm của QUIC trong quá trình kết nối - cancellation do bản chất UDP gây ra.

### code

```go
	signal_chan, command_chan, ctx_signal, ctx_client, ctx1, ctx2, stop, cancel1, cancel2 := step.Signal(passcode)
	defer cancel1()
	defer cancel2()
	log.Println("signal initiation done, connecting to the server....")

	client, err := step.InitClient(host, ctx_client, "caller", http3)

	if err != nil {
		log.Fatalf("error before connecting: %v \n", err)
	}
	
	/*-------------------------------------*/
func Signal(passcode string) (chan struct{}, chan string, context.Context, context.Context, context.Context, context.Context, context.CancelFunc, context.CancelFunc, context.CancelFunc) {
	signal_chan := make(chan struct{})

	command_chan := make(chan string)
	ctx_signal, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	md_data := metadata.Pairs("passcode", passcode)
	ctx_client := metadata.NewOutgoingContext(ctx_signal, md_data)
	ctx1, cancel1 := context.WithCancel(ctx_signal)
	// defer cancel1()
	ctx2, cancel2 := context.WithCancel(ctx_signal)
	// defer cancel2()
	return signal_chan, command_chan, ctx_signal, ctx_client, ctx1, ctx2, stop, cancel1, cancel2
}
/*-------------------------------------------*/
func InitClient(host string, ctx context.Context, passcode string, http3 bool) (pb.Calling_VoIPClient, error) {
	if http3 {
		return networking.InitV3(host, passcode, ctx)
	}
	return networking.InitV2(host, ctx)
}
```

## 2. server implementation.

Phần server, so với ý tưởng trước đây của tôi, giờ chỉ có nhiệm vụ quản lý kết nối và truyền tải dữ liệu nên có phần nhẹ hơn so với client, nhưng không có nghĩa là server không quan trọng, trái lại, việc tối ưu lại server cũng giúp hiệu suất chương trình kha khá do được viết lại đơn giản và bớt rườm rà hơn. Điểm mới của server lần này là chuyển từ array sang sync.Map, phù hợp hơn với concurrency vì có thể xóa hoặc thêm vào mà không phải lo về tình trạng deadlocks. Thêm nữa, nhờ có sự xuất hiện của HTTP/3 nên độ trễ được giảm xuống đáng kể, nhưng rủi ro về quá trình ngắt cũng lớn dần do sử dụng UDP, dẫu vậy, nó vẫn là một nguyên do khiến tôi quyết định thiết kế lại chương trình.

### 1.1. receiving side.

Phần receive thì khá đơn giản, với mỗi connection, buffered channel có nhiệm vụ thu thập các gói tin, việc kiểm tra loại gói tin, ở đây, hệ quả từ việc thay đổi protobuf, thì giờ tôi kiểm tra xem gói tin trả về có phải là tín hiệu hay không, nếu có thì sẽ giở luôn tín hiệu trả về và sẽ không xử lý thêm, còn ngược lại sẽ serialize rồi cho thẳng vào data_chan. Kèm theo đó, receive sẽ có cơ chế ngắt nếu như server không nhận được gì từ phía client.

### code for receiving side

```go
func receive(conn pb.Calling_VoIPServer, data_chan chan *types.Chunk, id uuid.UUID, signal chan struct{}) {
	defer close(signal)
	for {
		segment, err := conn.Recv()

		if err != nil {
			// log.Println("receiving side...")
			return
		}

		if segment.GetSignal() != nil {
			conn.Send(&pb.ServerRES{
				Response: &pb.ServerRES_Signal{
					Signal: &pb.ServerSignal{},
				},
			})
			continue
		}
		//for testing connection from client...
		/*
			if the request is signal -> return the signal as successully for the test
			else, just process the sound

			cuz, these could do <=> the connection is initiated!!!!!
		*/
		data := segment.GetMessage()
		c := &types.Chunk{
			ID:    id,
			Name:  data.GetName(),
			Chunk: data.GetChunk(),
		}

		data_chan <- c
	}
}
```

### 1.2. sending side.

Phần sending cũng tương tự, nhưng hoạt động ngược lại với receiving side. Lúc này, thay vì lấy dữ liệu rồi thả vào channel, giờ phải lấy từ channel rồi thả vào client. Do đây là phía gửi cho mỗi client, dẫn đến việc phải có cơ chế quản lý kết nối cho mỗi connection để đảm bảo hiệu suất chương trình tốt nhất.

### code for sending side.

```go
func send(conn pb.Calling_VoIPServer, data_chan chan *types.Chunk, id uuid.UUID, signal chan struct{}) {
	defer close(signal)
	for {
		data, ok := <-data_chan
		if !ok {
			return
		}
		conn.Send(
			&pb.ServerRES{
				Response: &pb.ServerRES_Message{
					Message: &pb.ServerMSG{
						Msg: &pb.ClientMSG{
							Chunk: data.Chunk,
							Name:  data.Name,
						},
						Id: id.String(),
					},
				},
			},
		)
	}
}

```

### 1.3. control mechanism

Bài toán ở đây, giờ được nâng lên nhiều lần, khi mà các client tương tác với nhau thì phải có một cơ chế quản lý kết nối, không giống như trước đó, khi mà các connection hoạt động riêng rẽ. Ở đây, phần quản lý kết nối được chia ra làm 2 loại, là quản lý trên mỗi kết nối và quản lý toàn bộ kết nối cũng như cơ chế ngắt.

### connection in - out

Trước khi đến phần kiểm soát kết nối, thì việc phản ứng mỗi khi các kết nối đến và đi rất quan trọng, thậm chí, việc kiểm soát quá trình đi ra đi vào của kết nối được coi là sống còn với các phiên bản trước, vốn sử dụng array vô cùng nhạy cảm với các thay đổi và liên tục phải sử dụng mutex cũng khiến chương trình vô cùng phức tạp.

### connection in

Danh sách việc cần làm:

1. Tạo ra một id định danh cho kết nối đó
2. Tạo ra một data channel với capactity là 1000 để tránh mất mát dữ liệu trong điều kiện mạng ngặt nghèo.
3. Nếu tạo thành công thì tạm ngưng một đoạn gói tin để thông báo xuất hiện client mới.

```go
func (s *Serv) In() (xuuid.UUID, error) {

	id := xuuid.NewV4()
	data := make(types.Conn, 1000)
	_, stored := s.Output.LoadOrStore(id, &data)
	if stored {
		return xuuid.Nil, errors.New("error in creating channel")
	}
	s.Incoming <- id
	return id, nil
}
```

### connection out

Một khi mà chúng out, sẽ có 2 dạng là chủ động và bị động, với chủ động thì sẽ báo cho server hủy bỏ kết nối và báo lại để sắp xếp lại các kết nối và các channel liên quan, còn ngược lại thì sẽ để cho quá trình đó diễn ra sau cùng.

```go
func (s *Serv) out(id xuuid.UUID, closed bool) {
	log.Printf("Delete for %v is called \n", id.String())
	defer log.Printf("%v is released \n", id.String())
	if !closed {
		s.Leaving <- id
		log.Println("Deletion requested!!!")
		return
	}

	log.Printf("Released %v!!!", id.String())
}

```

### control-per-connection

```go

func (s *Serv) VoIP(conn pb.Calling_VoIPServer) error {
	// id, err := s.In()
	// if err != nil {
	// 	return status.Error(codes.Internal, "create got problem...")
	// }
	id := conn.Context().Value(types.T("channel")).(uuid.UUID)
	res, ok := s.Output.Load(id)
	if !ok {
		return status.Error(codes.Internal, "losing channel...")
	}
	channel := *res.(*types.Conn)

	sig1 := make(chan struct{})
	sig2 := make(chan struct{})
	//receive the byte to send to the channel
	go receive(conn, s.Input, id, sig1)
	//get the data then send to the client
	go send(conn, channel, id, sig2)
	for {
		select {
		case <-sig1:
			log.Println("Disconnected....")
			s.out(id, false)
			return nil
		case <-sig2:
			log.Println("Forcing closed!!!!")
			s.out(id, true)
			return nil
		}
	}
}

```

Việc đầu tiên khi một connection được hình thành là phải có được id của kết nối đó, thì tôi chọn cách tạo nó trong interceptor, giờ việc chính là lấy nó ra và sử dụng những thứ được gán với nó, rất đơn giản. Tiếp đến, 2 goroutines xử lý công việc truyền tải và buffering dữ liệu, nhưng phải có cách để ngắt toàn bộ chương trình khi nó gặp lỗi, thì cơ chế cancellation được sử dụng với channel, chỉ cần 1 trong 2 quá trình bị lỗi thì ngay lập tức sẽ ngắt toàn bộ chương trình. Dĩ nhiên là sau khi mỗi kết nối bị hủy thì phải dọn thi thể, và hàm out được sử dụng như là cơ chế signaling vậy.

### multi-connection management and cancellation.

Nhưng, chỉ xử lý mỗi kết nối là chưa đủ, vì chúng hoạt động tương tác với nhau nên giờ phải xử lý đa kết nối, thêm cả quá trình ngắt bị động, ở đây là cancellation Ctrl+C.

Trong quá trình quản lý kết nối này có 4 trạng thái chính, bao gồm quá trình ngắt bị động với context (ctx), xử lý khi client đến và đi (incoming/leaving) và gửi giữa các gói tin.

### client incoming - leaving

```go
case x := <-serv.Incoming:
			log.Printf("%v is coming!!! \n", x.String())
			continue
		case x := <-serv.Leaving:
			log.Printf("%v is leaving!!! \n", x.String())
			res, ok := serv.Output.Load(x)
			if !ok {
				return
			}
			channel := *res.(*types.Conn)
			close(channel)
			serv.Output.Delete(x)
			<-interceptor.ConnLimit
```

### cancellation

```go
case <-ctx.Done():
			log.Println("cancellation is received, start to close....")
			return
```

### connections interaction

```go
case data, ok := <-serv.Input:
			if !ok {
				log.Println("Channel is forcibly closed")
				return
			}
			serv.Output.Range(func(key, value any) bool {
				channel := *value.(*types.Conn)
				id := key.(uuid.UUID)
				if !uuid.Equal(data.ID, id) {
					channel <- data
				}
				// log.Printf("We got data from: %v, %v", id, <-channel)
				return true
			})
```

Ở đây, do việc sử dụng sync.Map nên quá trình gửi chỉ cần dùng hàm Range, kiểm tra trùng lặp id rồi gửi đi cho channel gắn với từng client connections.

### 1.4. interceptor: connection limiting, simple authentication, channel finding.

Tại đây, tôi có diagram, và 3 loại interceptor hoạt động theo thứ tự, là connection limiting với việc sử dụng buffered channel để đếm connection, simple authentication để làm xác thực và cuối cùng là định danh rồi gói lại thông tin đưa về cho connection để quản lý.

### diagram

![Untitled Diagram.drawio(4).svg](diagram/Untitled_Diagram.drawio(4).svg)

### connection limiting

```go

var ConnLimit chan struct{}
var MaxConn = 10

func init() {
	ConnLimit = make(chan struct{}, MaxConn)
}

// func(srv interface{}, ss ServerStream, info *StreamServerInfo,handler StreamHandler) error
func Limiting(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// passcode := ss.Context().Value("passcode").(string)
	// log.Println(passcode)

	p, ok := peer.FromContext(ss.Context())
	log.Printf("%v is called... \n", info.FullMethod)
	if !ok {
		log.Println("Cannot get any context.....")
	}
	log.Printf("%v wants to connect....: ", p.Addr.String())

	a := MaxConn - len(ConnLimit)
	if a <= 0 {
		// log.Printf("remaining %v connections... \n", a)
		log.Println("Out of slots")
		return status.Error(codes.ResourceExhausted, "out of slots, cancelled!!!")
	}
	log.Println("Accepted!!!")
	ConnLimit <- struct{}{}
	a = 10 - len(ConnLimit)
	log.Printf("remaining %v connections... \n", a)
	handler(srv, ss)
	return nil
}

```

### channel finding

```go

// var ConnLimit chan struct{}

// func init() {
// 	ConnLimit = make(chan struct{}, 10)
// }

// func(srv interface{}, ss ServerStream, info *StreamServerInfo,handler StreamHandler) error
type wrapped struct {
	grpc.ServerStream
	id uuid.UUID
}

// for context....

func (w *wrapped) Context() context.Context {
	return context.WithValue(w.ServerStream.Context(), types.T("channel"), w.id)
}
func ChannelFinding(s *serv.Serv) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		id, err := s.In()
		if err != nil {
			<-ConnLimit
			return status.Error(codes.NotFound, "slot not found...")
		}
		log.Printf("slot %v is opened!!!", id)
		handler(srv, &wrapped{ss, id})
		return nil

	}
}

```

### simple authentication

```go

func AuthenClone(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		log.Println("does not have any context....")
	}
	passcode := md.Get("passcode")[0]
	if passcode != types.Passcode {
		log.Println("wrong passcode!!!!")
		<-ConnLimit
		return status.Error(codes.Unauthenticated, "out of slots, cancelled!!!")
	}
	// log.Println(passcode)
	handler(srv, ss)
	return nil
}

```

# iv. evaluation and bigger potential.

Vậy là dự án đến đây là hết, và có một vài điểm mà chương trình có thể phát triển thêm:

- I/O management at client: quản lý IO ở phía client để client chủ động hơn trong việc bật hay tắt tiếng, thay vì chỉ mỗi tắt thu âm.
- Hmmm, sẽ như thế nào nếu số lượng connection lên tới 100, 1000? Khi đó, bài toán trở thành tính toán cụm và một vài ý tưởng sẽ được đưa ra vào bài viết chính, bao gồm P2P, master-slave,…

# The end, a wonderful project!!!

Written by Bronstein, est2016.