package bigmodeltts

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestTTSClient(t *testing.T) {
	client, err := NewTTSWsClient(
		WithAppKey("xxxx"),
		WithAccessKey("xxxx"))

	if err != nil {
		t.Fatal(err)
	}

	sender, receiver, err := client.Connect(context.Background(), &AudioSettings{
		Format:     "m4a",
		SampleRate: 16000,
		Channel:    1,
		Speaker:    "xxx",
		ResourceID: "volc.megatts.default",
	})

	if err != nil {
		t.Fatal(err)
	}

	sessionID, err := sender.StartSession(context.Background(), "BidirectionalTTS")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(sessionID)

	text := []rune(`关于杭州六小龙现象，我来分析一下当前的情况。

让我从两个维度来分析一下杭州六小龙现象:

从行业角度看，我留意到六小龙涉及的赛道目前都处于较好的发展态势。从最新市场数据来看，相关的虚拟现实、机器视觉、云计算等行业都呈现持续多头信号。特别是虚拟现实行业已经连续18天保持多头，机器视觉行业也维持了8天的多头信号。同时，科技创新板块表现活跃，今天创业板指数上涨2.36%，显示市场对科技创新领域信心较足。

从技术实力看，这六家企业各有特色：
- 深度求索在AI大模型领域实现了突破，成本优势明显
- 宇树科技的机器人技术已达到全球领先水平
- 群核科技是全球最大的3D云设计平台之一
- 强脑科技是全球少数几家成功量产脑机接口的企业
- 游戏科学的《黑神话：悟空》展现了国产游戏的高水平
- 云深处科技在工业级机器人领域有独特优势

投资重点可以关注以下几个方向：

1. 核心技术产业链：主要是机器人和AI产业链的关键环节，包括高端芯片、传感器、减速器等核心零部件企业。目前半导体、专用机械等相关行业都出现了强势多头信号。

2. 应用场景拓展：六小龙的技术已经开始在多个领域落地，如工业自动化、智能设计、医疗康复等。这些下游应用领域的龙头企业值得关注。

3. 产业生态圈：与六小龙有深度合作的上市公司，以及在AI计算、数据服务等基础设施领域布局的企业。

需要注意的几个风险点：

1. 科技创新企业普遍估值较高，市场波动可能较大
2. 商业化进程存在不确定性
3. 技术迭代和人才竞争的压力
4. 全球科技竞争加剧的影响

总的来说，杭州六小龙代表了中国科技创新的最新成果，反映出我国在AI、机器人等前沿领域的突破。这个方向值得重点跟踪，但投资时要注意风险把控，选择具备核心竞争力的优质标的。`)

	for i := 0; i < len(text); i += 10 {
		data := ""
		if i+10 < len(text) {
			data = string(text[i : i+10])
		} else {
			data = string(text[i:])
		}
		fmt.Println(data)
		if strings.TrimSpace(data) == "" {
			continue
		}
		err = sender.SendText(context.Background(), sessionID, data)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := sender.FinishSession(context.Background(), sessionID); err != nil {
		t.Fatal(err)
	}

	dataChan, errChan := receiver.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.OpenFile("./aifrank.m4a", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	for {
		select {
		case data := <-dataChan:
			t.Log(len(data.Data))
			f.Write(data.Data)
		case err := <-errChan:
			t.Fatal(err)
			return
		}
	}
}
