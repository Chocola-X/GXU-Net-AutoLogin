package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"net/http"
	"net"
	"net/url"
	"time"
	"io"
)

const configFileName = "config.txt"

// Config 结构体保存配置
type Config struct {
	User        string
	Password    string
	NetType     string // 新增字段
	StudentMode bool

	// 路由器模式（当两者都非空时启用）
	RouterIP  string
	RouterMAC string
}

// loadConfig 加载或创建配置文件
func loadConfig() (*Config, error) {
	// 检查文件是否存在
	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		// 创建默认模板
		defaultContent := `# 校园网登陆脚本信息设置：（注意请不要改变格式）
# 用户名：（填写示例：User=1807210721）
User=
# 密码：（填写示例：Password=www.nekopara.uk）
Password=
# 运营商选择，留空选择校园网，如果需要选择运营商，电信填写telecom，联通填写unicom，移动填写cmcc
Net_Type=
# 是否开启学生上网时段模式？1为开启，0为关闭，开启后周一到周五0:00-6:00将不会尝试重连
Student_Mode=0
# 开启路由器登陆模式：
# 如果填写以下两个参数（均非空），则使用指定的路由器IP和MAC进行认证。
# 否则使用本机IP和MAC。
# 示例：
# Router_IP=172.16.6.6
# Router_MAC=36:88:8A:99:A4:CC
Router_IP=
Router_MAC=
`

			err = os.WriteFile(configFileName, []byte(defaultContent), 0644)
			if err != nil {
				return nil, fmt.Errorf("无法创建配置文件: %v", err)
			}
			return nil, fmt.Errorf("配置文件 '%s' 已创建，请先填写上网信息后重新运行程序", configFileName)
	}

	// 读取并解析
	content, err := os.ReadFile(configFileName)
	if err != nil {
		return nil, fmt.Errorf("无法读取配置文件: %v", err)
	}

	cfg := &Config{}
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 按第一个 '=' 分割（避免密码含等号出错）
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // 格式错误，跳过
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
			case "User":
				cfg.User = value
			case "Password":
				cfg.Password = value
			case "Net_Type":
				cfg.NetType = value // 新增这一行
			case "Student_Mode":
				cfg.StudentMode = (value == "1")
			case "Router_IP":
				cfg.RouterIP = value
			case "Router_MAC":
				cfg.RouterMAC = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 基础校验
	if cfg.User == "" || cfg.Password == "" {
		return nil, fmt.Errorf("请在 '%s' 中填写用户名和密码", configFileName)
	}
	// 在 loadConfig 函数中，解析配置后添加：
	if cfg.NetType != "" {
		// 检查是否是合法的运营商
		valid := false
		switch strings.ToLower(cfg.NetType) {
			case "telecom", "unicom", "cmcc":
				valid = true
		}

		if !valid {
			return nil, fmt.Errorf("错误：运营商类型必须为空、telecom、unicom或cmcc（不区分大小写），当前值: %s", cfg.NetType)
		}
	}

	return cfg, nil
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String(), nil
}

func getMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		mac := iface.HardwareAddr.String()
		if mac == "" {
			continue
		}

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
				return mac, nil
			}
		}
	}
	return "", fmt.Errorf("no active network interface with MAC found")
}

func isNetworkOK() bool {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := client.Get("http://connect.rom.miui.com/generate_204")
	if err != nil {
		return false // 网络不通 / DNS 故障 / 超时
	}
	defer resp.Body.Close()

	return resp.StatusCode == 204
}

func login(cfg *Config, ip, mac string) {
	// 格式化 MAC：去掉冒号，转小写（适配你 bash 脚本的行为）
	cleanMAC := strings.ReplaceAll(strings.ToLower(mac), ":", "")

	userAccount := cfg.User
	if cfg.NetType != "" {
		userAccount = cfg.User + "@" + cfg.NetType
	}

	params := url.Values{
		"callback":       {"dr1003"},
		"login_method":   {"1"},
		"user_account":   {userAccount},
		"user_password":  {cfg.Password},
		"wlan_user_ip":   {ip},
		"wlan_user_mac":  {cleanMAC},
		"wlan_user_ipv6": {""},
		"wlan_ac_ip":     {""},
		"wlan_ac_name":   {""},
		"jsVersion":      {"4.2.1"},
		"terminal_type":  {"1"},
		"lang":           {"zh-cn"},
		"v":              {"5574"},
	}

	loginURL := "http://172.17.0.2:801/eportal/portal/login?" + params.Encode()

	resp, err := http.Get(loginURL)
	if err != nil {
		fmt.Printf("❌ 登录请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 读取并打印响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ 读取响应体失败: %v\n", err)
		return
	}
	bodyStr := string(body)

	fmt.Printf("✅ 已发送登录请求（HTTP状态码: %d）\n", resp.StatusCode)
	fmt.Printf("响应内容: %s\n", bodyStr)
}

func shouldSkipLogin(cfg *Config) bool {
	if !cfg.StudentMode {
		return false
	}

	now := time.Now()
	weekday := now.Weekday() // Sunday = 0, Monday = 1, ..., Friday = 5
	hour := now.Hour()

	// 周一到周五（1~5），且 0:00 ~ 5:59
	if weekday >= time.Monday && weekday <= time.Friday && hour >= 0 && hour < 6 {
		fmt.Println("🌙 学生模式：当前为禁网时段，暂停重连")
		return true
	}

	return false
}

func getLoginInfo(cfg *Config) (ip, mac string, err error) {
	// 如果启用了路由器模式（两个字段都非空）
	if cfg.RouterIP != "" && cfg.RouterMAC != "" {
		fmt.Println("🌐 使用路由器模式进行认证")
		return cfg.RouterIP, cfg.RouterMAC, nil
	}

	// 否则使用本机信息
	fmt.Println("💻 使用本机模式进行认证")
	ip, err = getLocalIP()
	if err != nil {
		return "", "", fmt.Errorf("获取本机IP失败: %w", err)
	}
	mac, err = getMACAddress()
	if err != nil {
		return "", "", fmt.Errorf("获取本机MAC失败: %w", err)
	}
	return ip, mac, nil
}

func main() {
	fmt.Printf("🚀广西大学校园网自动登陆程序 By：GTX690战术核显卡导弹（www.nekopara.uk）\n")
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("❌ 错误:", err)
		fmt.Println("💡 请编辑 config.txt 后重新运行本程序。")
		os.Exit(1)
	}

	fmt.Printf("✅ 配置加载成功！\n")
	fmt.Printf("用户: %s\n", cfg.User)
	fmt.Printf("密码: %s\n", cfg.Password)
	fmt.Printf("运营商: %s\n", cfg.NetType) // 新增这一行
	fmt.Printf("学生模式: %t\n", cfg.StudentMode)
	if cfg.RouterIP != "" && cfg.RouterMAC != "" {
		fmt.Printf("路由器模式: IP=%s, MAC=%s\n", cfg.RouterIP, cfg.RouterMAC)
	}

	// 获取用于登录的 IP 和 MAC（自动判断模式）
	ip, mac, err := getLoginInfo(cfg)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 守护进程启动：认证IP=%s | 认证MAC=%s\n", ip, mac)

	// 主循环
	for {
		if shouldSkipLogin(cfg) {
			time.Sleep(1 * time.Second)
			continue
		}

		if !isNetworkOK() {
			fmt.Println("⚠️ 检测到断网，正在重新登录...")
			login(cfg, ip, mac)
		}

		time.Sleep(1 * time.Second)
	}
}
