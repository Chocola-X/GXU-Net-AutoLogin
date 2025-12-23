# 🌐 GXU-Net-AutoLogin

> 广西大学校园网自动登录 & 断网重连守护程序  
> 支持学生账号禁网时段智能跳过 · 路由器模式 · 运营商选择 · 配置文件/命令行双模式

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
![Go Version](https://img.shields.io/badge/Go-1.20%2B-informational)
![Platform](https://img.shields.io/badge/Platform-Linux%20|%20Windows-lightgrey)

广西大学校园网在高峰期或夜间（尤其是周一至周五 0:00–6:00）常会断连，手动重登既麻烦又影响挂机任务。本程序使用 **Go 语言** 编写，轻量高效，可实现：
- ✅ **断网自动检测并重连**
- ✅ **支持校园网 + 三大运营商（电信/联通/移动）**
- ✅ **学生模式：禁网时段自动暂停重连**
- ✅ **路由器模式：指定 IP/MAC 登录（适配宿舍共享上网）**
- ✅ **配置文件 or 命令行参数，灵活部署**

---

## 📦 快速开始

### 方法一：使用预编译二进制

1. 从 [Release](https://github.com/Chocola-X/GXU-Net-AutoLogin/releases) 下载对应的二进制文件
2. 运行 `GXU_Net_AutoLogin`
3. 首次运行会生成 `config.txt`，按提示填写账号密码
4. 再次运行即可后台守护

### 方法二：使用包管理器安装

Arch Linux 用户可从 [AUR](https://aur.archlinux.org/packages/gxu-net-autologin) 安装：

```bash
[yay/paru] -S gxu-net-autologin
```

配置文件路径为 `/etc/gxu-net-autologin/config.txt`。

### 方法三：从源码编译（推荐 Linux 用户）

确保已安装 [Go 1.20+](https://golang.org/dl/)

```bash
git clone https://github.com/Chocola-X/GXU-Net-AutoLogin.git
cd GXU-Net-AutoLogin
go build -ldflags="-s -w" -o GXU_Net_AutoLogin main.go
./GXU_Net_AutoLogin
```

首次运行将自动生成 `config.txt`，编辑后重新运行即可。

---

## ⚙️ 配置说明

### 方式 1：配置文件（`config.txt`）

程序首次运行会自动生成模板，设置示例如下：

```ini
# 校园网登录脚本信息设置：（注意请不要改变格式）
User=1807210721
Password=your_password_here
# 运营商：留空=校园网，telecom=电信，unicom=联通，cmcc=移动
Net_Type=cmcc
# 学生模式：1=开启（周一至五 0:00-6:00 不重连），0=关闭
Student_Mode=1
# 路由器模式（两者需同时填写才生效）：
Router_IP=172.16.6.6
Router_MAC=36:88:8A:99:A4:CC
```

### 方式 2：命令行参数（适合服务部署）

```bash
# 基础用法
./GXU_Net_AutoLogin -user 1807210721 -passwd your_password

# 完整示例（含运营商+学生模式+路由器）
./GXU_Net_AutoLogin \
  -user 1807210721 \
  -passwd mypassword \
  -nettype cmcc \
  -studentmode \
  -ip 172.16.6.6 \
  -mac 36:88:8A:99:A4:CC
```

> 💡 **注意**：`-ip` 和 `-mac` 必须**同时提供**，否则视为无效。

查看全部参数：
```bash
./GXU_Net_AutoLogin -help
```

---

## 🔧 技术原理

程序通过向广西大学认证服务器发送标准 ePortal 登录请求实现联网：

```
GET http://172.17.0.2:801/eportal/portal/login?
  callback=dr1003&
  login_method=1&
  user_account=账号[@运营商]&
  user_password=密码&
  wlan_user_ip=终端IP&
  wlan_user_mac=设备MAC（无冒号小写）&
  ...
```

- **网络检测**：每秒请求 `http://connect.rom.miui.com/generate_204`（返回 204 表示联网正常）
- **MAC 获取**：自动读取本机活跃网卡 MAC（非 `00:00:00:00:00:00`，避免频繁掉线）
- **时段控制**：学生模式下，周一至五 0:00–5:59 自动暂停重连，避免无效请求

> 参考官方文档：[Linux系统宽带客户端-2024.12.30日后使用](https://net.gxu.edu.cn/info/1360/2293.htm)

---

## 🛠️ 功能特性

| 特性 | 说明 |
|------|------|
| ✅ 自动重连 | 网络中断后 1 秒内自动尝试登录 |
| ✅ 多运营商支持 | `telecom` / `unicom` / `cmcc` |
| ✅ 路由器模式 | 指定任意 IP/MAC 登录|
| ✅ 学生禁网时段 | 智能跳过无效重连 |
| ✅ 极低资源占用 | Go 编译为静态二进制，内存 < 10MB |
| ✅ 跨平台 | Linux / Windows 均可运行 |

---

## 📜 许可证

本项目采用 [GNU Affero General Public License v3.0 (AGPL-3.0)](LICENSE) 开源协议。

> 如果你修改了代码并用于网络服务（如部署为公共代理），**必须公开修改后的源码**。

---

## 🙏 鸣谢

- 广西大学信息网络中心 提供的 [Linux 登录接口](https://net.gxu.edu.cn/info/1360/2293.htm)
- MIUI 的 `generate_204` 网络探测机制

---

> **作者**：GTX690战术核显卡导弹（[nekopara.uk](https://www.nekopara.uk)）  
> **仓库**：[github.com/Chocola-X/GXU-Net-AutoLogin](https://github.com/Chocola-X/GXU-Net-AutoLogin)

🚀 **Enjoy your stable campus network!**
