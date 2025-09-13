package util

import (
    "bufio"
    "flag"
    "os"
    "path/filepath"
    "strings"
)

// LoadDotEnv 读取 .env 文件，按 KEY=VALUE 解析并设置到环境变量。
// 若某个 key 在环境中已存在，则保持现有值（不覆盖）。
func LoadDotEnv(files ...string) error {
    for _, f := range files {
        if f == "" {
            continue
        }
        path := f
        if !filepath.IsAbs(path) {
            // 相对路径基于当前工作目录
            cwd, _ := os.Getwd()
            path = filepath.Join(cwd, f)
        }
        fp, err := os.Open(path)
        if err != nil {
            // 文件不存在则跳过
            continue
        }
        scanner := bufio.NewScanner(fp)
        for scanner.Scan() {
            line := strings.TrimSpace(scanner.Text())
            if line == "" || strings.HasPrefix(line, "#") {
                continue
            }
            if strings.HasPrefix(line, "export ") {
                line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
            }
            kv := strings.SplitN(line, "=", 2)
            if len(kv) != 2 {
                continue
            }
            key := strings.TrimSpace(kv[0])
            val := strings.TrimSpace(kv[1])
            // 去掉包裹引号
            if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) || (strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
                val = val[1:len(val)-1]
            }
            if os.Getenv(key) == "" {
                _ = os.Setenv(key, val)
            }
        }
        fp.Close()
    }
    return nil
}

// FlagSpec 用于描述一个 flag，同时绑定到环境变量
type FlagSpec struct {
    Name    string // 例如 "listen"
    EnvKey  string // 例如 "LISTEN_ADDR"
    Default string // 默认值（在 env 缺省时使用）
    Usage   string // 用法描述
}

// ParseFlagsWithEnv 使用自定义 FlagSet 解析命令行，并与环境变量合并。
// 返回解析后的键值对（key 为 Name）。命令行优先级最高，其次环境变量，最后 Default。
func ParseFlagsWithEnv(specs ...FlagSpec) map[string]string {
    fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
    out := map[string]*string{}
    for _, s := range specs {
        def := os.Getenv(s.EnvKey)
        if def == "" {
            def = s.Default
        }
        out[s.Name] = fs.String(s.Name, def, s.Usage)
    }
    _ = fs.Parse(os.Args[1:])
    ret := map[string]string{}
    for _, s := range specs {
        ret[s.Name] = *out[s.Name]
    }
    return ret
}

