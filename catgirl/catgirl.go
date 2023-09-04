package catgirl

import (
	"encoding/json"
	"fmt"
	"github.com/FloatTech/floatbox/binary"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	cdb *catgirldb
	// datapath string
	// predictRe = regexp.MustCompile(`{"steps".+?}`)
	// 参考host http://91.217.139.190:5010 http://91.216.169.75:5010
	// aipaintImg2ImgURL = "/got_image2image?token=%v&tags=%v"
	cfg = newServerConfig("data/catgirl/config.json")
)

const (
	ua         = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36"
	drawURL    = "https://ai.huashi6.com/aiapi/draw"
	infoURL    = "https://rt.huashi6.com/front/ai/paintRecord/detail?id="
	processURL = "https://ai.huashi6.com/aiapi/progress"
)

func init() { // 插件主体

	engine := control.Register("catgirl", &ctrl.Options[*zero.Ctx]{
		DisableOnDefault: false,
		Brief:            "猫娘",
		Help: "- [ 添加猫娘 | 成为猫娘 ] 名字 身高 体重 年龄 特点\n" +
			"- 查看本群猫娘\n" +
			"- 猫娘绘图 [猫娘名字] [猫娘行为]\n" +
			"- 设置猫娘cookie [cookie]\n" +
			"- 查看猫娘绘图配置\n",
		PrivateDataFolder: "catgirl",
	})

	// 加载数据库
	dbpath := engine.DataFolder()
	dbfile := dbpath + "cat.db"
	cdb = initializeCat(dbfile)

	if file.IsNotExist(cfg.file) {
		s := serverConfig{}
		data, err := json.Marshal(s)
		if err != nil {
			panic(err)
		}
		err = os.WriteFile(cfg.file, data, 0666)
		if err != nil {
			panic(err)
		}
	}

	engine.OnRegex(`^[添|成][加|为]猫娘\s(.*[^\s$])\s([0-9]*)\s([0-9]*)\s([0-9]*)\s(.+)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid == 0 {
				ctx.SendChain(message.Text("ERROR: IS NOT A GROUP"))
			}
			regexMatched := ctx.State["regex_matched"].([]string)
			tall, _ := strconv.ParseInt(regexMatched[2], 10, 64)
			weight, _ := strconv.ParseInt(regexMatched[3], 10, 64)
			age, _ := strconv.ParseInt(regexMatched[4], 10, 64)
			if err := addCatGirl(gid, regexMatched[1], tall, weight, age, regexMatched[5]); err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("已添加" + regexMatched[1] + "猫娘"))
		})

	engine.OnPrefixGroup([]string{`查看本群猫娘`, `查看猫娘`}).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid == 0 {
				ctx.SendChain(message.Text("ERROR: IS NOT A GROUP"))
			}
			cl := cdb.getAllGroupCatGirl(gid)
			msg := "----------------本群猫娘列表----------------"
			for _, v := range cl {
				msg += fmt.Sprintf("\n名字:%s 身高:%d 体重:%d 年龄:%d 特点:%s", v.Name, v.Tall, v.Weight, v.Age, v.Character)
			}
			data, err := text.RenderToBase64(msg, text.FontFile, 600, 20)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			if id := ctx.SendChain(message.Image("base64://" + binary.BytesToString(data))); id.ID() == 0 {
				ctx.SendChain(message.Text("ERROR: 可能被风控了"))
			}

		})

	engine.OnRegex(`^设置猫娘cookie\s(.*)$`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			regexMatched := ctx.State["regex_matched"].([]string)
			err := cfg.load()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			err = cfg.update(regexMatched[1])
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("成功设置\ncookie: ", cfg.Cookie))
		})

	engine.OnFullMatch(`查看猫娘绘图配置`, zero.SuperUserPermission).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			err := cfg.load()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			ctx.SendChain(message.Text("cookie: ", cfg.Cookie))
		})

	engine.OnRegex(`^猫娘绘图\s(.*[^\s$])\s(.*)$`).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			gid := ctx.Event.GroupID
			if gid == 0 {
				gid = -ctx.Event.UserID
				//ctx.SendChain(message.Text("ERROR: IS NOT A GROUP"))
			}
			regexMatched := ctx.State["regex_matched"].([]string)
			err := cfg.load()
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			params := getCatPrompt(gid, regexMatched[1]) + "," + regexMatched[2]
			paintAiImg(ctx, params)
		})
}

// add cat
func addCatGirl(groupId int64, name string, tall int64, weight int64, age int64, character string) (err error) {
	bpMap := map[string]any{
		"group_id":  groupId,
		"name":      name,
		"tall":      tall,
		"weight":    weight,
		"age":       age,
		"character": character,
	}
	return cdb.insertOrUpdateCatGirl(bpMap)
}

func paintAiImg(ctx *zero.Ctx, param string) {
	prompt := "猫娘,精致的五官," + param
	cli := web.NewDefaultClient()
	fmt.Println(prompt)
	ctx.SendChain(message.Text("开始绘图中..."))
	data, err := web.RequestDataWithHeaders(cli, drawURL, "POST", func(r *http.Request) error {
		r.Header.Set("User-Agent", ua)
		r.Header.Set("cookie", cfg.Cookie)
		r.Header.Set("Referer", "https://www.huashi6.com/ai/draw")
		r.Header.Set("Origin", "https://www.huashi6.com")
		r.Header.Set("Content-Type", "application/json;charset=UTF-8")
		return nil
	}, setDrawData(prompt))
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	var res DataRes
	err = json.Unmarshal(data, &res)
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	go func(cli *http.Client, dr AllData) {
		for {
			data, err = web.RequestDataWithHeaders(cli, processURL, "POST", func(r *http.Request) error {
				r.Header.Set("User-Agent", ua)
				r.Header.Set("cookie", cfg.Cookie)
				r.Header.Set("Referer", "https://www.huashi6.com/ai/draw")
				r.Header.Set("Origin", "https://www.huashi6.com")
				r.Header.Set("Content-Type", "application/json;charset=UTF-8")
				return nil
			}, setProcessData(dr.PaintingSign))
			if err != nil {
				return
			}

			err = json.Unmarshal(data, &res)
			if err != nil {
				ctx.SendChain(message.Text("ERROR: ", err))
				return
			}
			pr := res.Data
			if pr.State == "success" {
				sendAiImg(ctx, cli, dr.Id)
				return
			}
			if pr.State == "fail" {
				ctx.SendChain(message.Text("绘图fail,cookie失效"))
				return
			}
			fmt.Println(pr.Progress, pr.State)
			time.Sleep(2 * time.Second)
		}
	}(cli, res.Data)

}

func sendAiImg(ctx *zero.Ctx, cli *http.Client, id int) {
	data, err := web.RequestDataWithHeaders(cli, infoURL+strconv.Itoa(id), "GET", func(r *http.Request) error {
		r.Header.Set("User-Agent", ua)
		r.Header.Set("cookie", cfg.Cookie)
		r.Header.Set("Referer", "https://www.huashi6.com/ai/draw")
		r.Header.Set("Origin", "https://www.huashi6.com")
		r.Header.Set("Content-Type", "application/json;charset=UTF-8")
		return nil
	}, nil)
	var res DataRes
	err = json.Unmarshal(data, &res)
	ir := res.Data
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err)) //bug
		return
	}

	img := ir.ImageUrl
	if err != nil {
		ctx.SendChain(message.Text("ERROR: ", err))
		return
	}
	ctx.SendChain(message.Image(img))

}

func getCatPrompt(gid int64, name string) (res string) {
	cat := cdb.getCatGirlByName(gid, name)
	res = "猫娘,高画质,多细节,精美绝伦的五官,面容精致的,"
	if cat.Age < 22 {
		res += "妹妹,"
	} else {
		res += "姐姐,"
	}
	if cat.Tall < 170 || cat.Weight < 60 {
		res += "娇小,"
	} else {
		res += "丰满,"
	}
	return res + cat.Character
}
