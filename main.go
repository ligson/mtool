package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"unsafe"
)

const version = "0.2.0"

type OutputFormat string

const (
	FormatTable OutputFormat = "table"
	FormatJSON  OutputFormat = "json"
	FormatPlain OutputFormat = "plain"
	FormatCSV   OutputFormat = "csv"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "fan":
		cmdFan(os.Args[2:])
	case "temp":
		cmdTemp(os.Args[2:])
	case "all":
		cmdAll(os.Args[2:])
	case "power":
		cmdPower()
	case "diag":
		cmdDiag()
	case "version", "-v", "--version":
		fmt.Println("mtool version", version)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`mtool - Mac sensor monitor (M-series compatible)

Usage:
  mtool <command> [options]

Commands:
  fan     Show fan speeds
  temp    Show temperature sensors
  all     Show all sensor data
  power   Show power/thermal data via powermetrics
  diag    Show SMC key diagnostics
  version Show version

Options (for temp, fan, all):
  -f, --format=<fmt>   Output format: table (default), json, plain, csv
  -g, --group          Group by sensor type and show averages
  -k, --key=<key>      Show only specific sensor key (e.g. Tp01)
  -t, --type=<type>    Show only specific sensor type: cpu, gpu, soc, battery, ambient, other

Examples:
  mtool temp                                # Table format, all sensors
  mtool temp -f json                        # JSON format
  mtool temp -f plain                       # Plain text (values only)
  mtool temp -g                             # Grouped averages (table)
  mtool temp -t cpu                         # CPU sensors only, grouped
  mtool temp -t cpu -f plain                # CPU average temperature, plain number
  mtool temp --type=gpu                     # GPU sensors (using = format)
  mtool temp -k Tp01                        # Single key
  mtool fan -f json
  mtool all -f csv
  mtool all -t soc -f json                  # SoC sensors in JSON`)
}

func openSMC() *SMC {
	smc := &SMC{}
	if err := smc.Open(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: SMC unavailable: %v\n", err)
		return nil
	}
	return smc
}

// parseCommonFlags parses format, group, key, and type flags
func parseCommonFlags(args []string) (format OutputFormat, group bool, key string, sensorType string) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String("format", "table", "")
	fs.String("f", "table", "")
	fs.Bool("group", false, "")
	fs.Bool("g", false, "")
	fs.String("key", "", "")
	fs.String("k", "", "")
	fs.String("type", "", "")
	fs.String("t", "", "")
	fs.Usage = func() {}
	fs.Parse(args)

	fmtStr := fs.Lookup("format").Value.String()
	if fmtStr == "table" {
		fmtStr = fs.Lookup("f").Value.String()
	}
	format = OutputFormat(fmtStr)

	groupVal := fs.Lookup("group").Value.String() == "true"
	if !groupVal {
		groupVal = fs.Lookup("g").Value.String() == "true"
	}
	group = groupVal

	key = fs.Lookup("key").Value.String()
	if key == "" {
		key = fs.Lookup("k").Value.String()
	}

	sensorType = fs.Lookup("type").Value.String()
	if sensorType == "" {
		sensorType = fs.Lookup("t").Value.String()
	}
	// Normalize type to uppercase (CPU, GPU, etc.)
	if sensorType != "" {
		sensorType = strings.ToUpper(sensorType)
	}

	return
}

func cmdFan(args []string) {
	format, _, _, _ := parseCommonFlags(args)

	smc := openSMC()
	if smc == nil {
		os.Exit(1)
	}
	defer smc.Close()

	count := smc.FanCount()
	if count == 0 {
		if format == FormatJSON {
			fmt.Println("[]")
		} else {
			fmt.Println("No fans detected")
		}
		return
	}

	var fans []map[string]interface{}
	for i := 0; i < count; i++ {
		actual, min, max, target := smc.FanRPM(i)
		fans = append(fans, map[string]interface{}{
			"id":     i,
			"actual": actual,
			"min":    min,
			"max":    max,
			"target": target,
		})
	}

	switch format {
	case FormatJSON:
		b, _ := json.MarshalIndent(fans, "", "  ")
		fmt.Println(string(b))
	case FormatPlain:
		for _, f := range fans {
			fmt.Printf("%.0f\n", f["actual"])
		}
	case FormatCSV:
		fmt.Println("id,actual,min,max,target")
		for _, f := range fans {
			fmt.Printf("%d,%.0f,%.0f,%.0f,%.0f\n", f["id"], f["actual"], f["min"], f["max"], f["target"])
		}
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "FAN\tACTUAL (RPM)\tMIN (RPM)\tMAX (RPM)\tTARGET (RPM)")
		fmt.Fprintln(w, "---\t------------\t---------\t---------\t-----------")
		for _, f := range fans {
			fmt.Fprintf(w, "Fan %d\t%.0f\t%.0f\t%.0f\t%.0f\n",
				f["id"], f["actual"], f["min"], f["max"], f["target"])
		}
		w.Flush()
	}
}

func cmdTemp(args []string) {
	format, group, key, sensorType := parseCommonFlags(args)

	smc := openSMC()
	if smc == nil {
		os.Exit(1)
	}
	defer smc.Close()

	sensors := smc.TemperatureSensors()
	if len(sensors) == 0 {
		if format == FormatJSON {
			fmt.Println("[]")
		} else {
			fmt.Println("No temperature sensors found")
		}
		return
	}

	// Filter by key if specified
	if key != "" {
		var filtered []TempSensor
		for _, s := range sensors {
			if s.Key == key {
				filtered = append(filtered, s)
			}
		}
		sensors = filtered
		if len(sensors) == 0 {
			fmt.Printf("Key %s not found\n", key)
			return
		}
	}

	// Filter by type if specified (e.g., --type cpu)
	if sensorType != "" {
		var filtered []TempSensor
		for _, s := range sensors {
			var category string
			if strings.HasPrefix(s.Key, "Tp") {
				category = "CPU"
			} else if strings.HasPrefix(s.Key, "Tg") {
				category = "GPU"
			} else if strings.HasPrefix(s.Key, "Te") {
				category = "SOC"
			} else if strings.HasPrefix(s.Key, "TB") {
				category = "BATTERY"
			} else if strings.HasPrefix(s.Key, "Ta") {
				category = "AMBIENT"
			} else if strings.HasPrefix(s.Key, "T") {
				category = "OTHER"
			} else {
				category = "UNKNOWN"
			}
			if category == sensorType {
				filtered = append(filtered, s)
			}
		}
		sensors = filtered
		if len(sensors) == 0 {
			fmt.Printf("Type %s not found\n", sensorType)
			return
		}
		// When filtering by type, auto-enable group mode for consistency
		group = true
	}

	if group {
		outputGrouped(sensors, format, sensorType)
	} else {
		outputSensors(sensors, format)
	}
}

func outputSensors(sensors []TempSensor, format OutputFormat) {
	var data []map[string]interface{}
	for _, s := range sensors {
		data = append(data, map[string]interface{}{
			"name":    s.Name,
			"key":     s.Key,
			"celsius": s.Celsius,
		})
	}

	switch format {
	case FormatJSON:
		b, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(b))
	case FormatPlain:
		for _, d := range data {
			fmt.Printf("%.1f\n", d["celsius"])
		}
	case FormatCSV:
		fmt.Println("name,key,celsius")
		for _, d := range data {
			fmt.Printf("%s,%s,%.1f\n", d["name"], d["key"], d["celsius"])
		}
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SENSOR\tKEY\tTEMPERATURE")
		fmt.Fprintln(w, "------\t---\t-----------")
		for _, s := range sensors {
			bar := tempBar(s.Celsius)
			fmt.Fprintf(w, "%s\t%s\t%.1f °C  %s\n", s.Name, s.Key, s.Celsius, bar)
		}
		w.Flush()
	}
}

func outputGrouped(sensors []TempSensor, format OutputFormat, filterType string) {
	groups := groupSensors(sensors)

	// If a specific type is filtered, show only that group
	if filterType != "" {
		for _, g := range groups {
			if g["group"].(string) == filterType {
				groups = []map[string]interface{}{g}
				break
			}
		}
	}

	switch format {
	case FormatJSON:
		b, _ := json.MarshalIndent(groups, "", "  ")
		fmt.Println(string(b))
	case FormatPlain:
		for _, g := range groups {
			fmt.Printf("%.1f\n", g["avg"])
		}
	case FormatCSV:
		fmt.Println("group,avg,count")
		for _, g := range groups {
			fmt.Printf("%s,%.1f,%d\n", g["group"], g["avg"], g["count"])
		}
	default:
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "GROUP\tAVG (°C)\tCOUNT\tDETAILS")
		fmt.Fprintln(w, "-----\t--------\t-----\t-------")
		for _, g := range groups {
			details := g["details"].(string)
			fmt.Fprintf(w, "%s\t%.1f\t%d\t%s\n", g["group"], g["avg"], g["count"], details)
		}
		w.Flush()
	}
}

// groupSensors groups sensors by type and calculates averages
func groupSensors(sensors []TempSensor) []map[string]interface{} {
	groups := make(map[string][]TempSensor)

	for _, s := range sensors {
		var category string
		if strings.HasPrefix(s.Key, "Tp") {
			category = "CPU"
		} else if strings.HasPrefix(s.Key, "Tg") {
			category = "GPU"
		} else if strings.HasPrefix(s.Key, "Te") {
			category = "SOC"
		} else if strings.HasPrefix(s.Key, "TB") {
			category = "BATTERY"
		} else if strings.HasPrefix(s.Key, "Ta") {
			category = "AMBIENT"
		} else if strings.HasPrefix(s.Key, "T") {
			category = "OTHER"
		} else {
			category = "UNKNOWN"
		}
		groups[category] = append(groups[category], s)
	}

	var result []map[string]interface{}
	order := []string{"CPU", "GPU", "SOC", "BATTERY", "AMBIENT", "OTHER", "UNKNOWN"}

	for _, cat := range order {
		if sensorList, ok := groups[cat]; ok && len(sensorList) > 0 {
			var sum float64
			var details strings.Builder
			for i, s := range sensorList {
				sum += s.Celsius
				if i > 0 {
					details.WriteString(", ")
				}
				details.WriteString(fmt.Sprintf("%s %.1f°C", s.Key, s.Celsius))
			}
			avg := sum / float64(len(sensorList))
			result = append(result, map[string]interface{}{
				"group":   cat,
				"avg":     avg,
				"count":   len(sensorList),
				"details": details.String(),
			})
		}
	}
	return result
}

func cmdAll(args []string) {
	format, group, _, sensorType := parseCommonFlags(args)

	// For JSON/CSV formats, combine everything
	if format == FormatJSON || format == FormatCSV || format == FormatPlain {
		smc := openSMC()
		if smc == nil {
			os.Exit(1)
		}
		defer smc.Close()

		sensors := smc.TemperatureSensors()
		if format == FormatJSON {
			data := map[string]interface{}{
				"temperatures": sensors,
			}
			count := smc.FanCount()
			if count > 0 {
				var fans []map[string]interface{}
				for i := 0; i < count; i++ {
					actual, min, max, target := smc.FanRPM(i)
					fans = append(fans, map[string]interface{}{
						"id": i, "actual": actual, "min": min, "max": max, "target": target,
					})
				}
				data["fans"] = fans
			}
			b, _ := json.MarshalIndent(data, "", "  ")
			fmt.Println(string(b))
		} else {
			if group || sensorType != "" {
				outputGrouped(sensors, format, sensorType)
			} else {
				outputSensors(sensors, format)
			}
		}
		return
	}

	// Table format (default)
	smc := openSMC()
	if smc != nil {
		defer smc.Close()

		// Fans
		count := smc.FanCount()
		printSection("FANS (SMC)")
		if count == 0 {
			fmt.Println("  No fans (passive cooling or unsupported)")
		} else {
			w := tabwriter.NewWriter(os.Stdout, 4, 0, 2, ' ', 0)
			fmt.Fprintln(w, "  FAN\tACTUAL\tMIN\tMAX\tTARGET")
			for i := 0; i < count; i++ {
				actual, min, max, target := smc.FanRPM(i)
				fmt.Fprintf(w, "  Fan %d\t%.0f RPM\t%.0f RPM\t%.0f RPM\t%.0f RPM\n",
					i, actual, min, max, target)
			}
			w.Flush()
		}

		fmt.Println()
		// Temperatures
		printSection("TEMPERATURES (SMC)")
		sensors := smc.TemperatureSensors()
		if len(sensors) == 0 {
			fmt.Println("  No SMC temperature keys found")
		} else {
			if group || sensorType != "" {
				groupList := groupSensors(sensors)
				// Filter by type if specified
				if sensorType != "" {
					for _, g := range groupList {
						if g["group"].(string) == sensorType {
							groupList = []map[string]interface{}{g}
							break
						}
					}
				}
				w := tabwriter.NewWriter(os.Stdout, 4, 0, 2, ' ', 0)
				fmt.Fprintln(w, "  GROUP\tAVG (°C)\tCOUNT")
				for _, g := range groupList {
					fmt.Fprintf(w, "  %s\t%.1f\t%d\n", g["group"], g["avg"], g["count"])
				}
				w.Flush()
			} else {
				w := tabwriter.NewWriter(os.Stdout, 4, 0, 2, ' ', 0)
				for _, s := range sensors {
					bar := tempBar(s.Celsius)
					fmt.Fprintf(w, "  %-28s\t[%s]\t%.1f °C  %s\n", s.Name, s.Key, s.Celsius, bar)
				}
				w.Flush()
			}
		}
		fmt.Println()
	}

	// Powermetrics
	printSection("POWER & THERMAL (powermetrics)")
	pm, err := RunPowermetrics()
	if err != nil {
		fmt.Println("  (unavailable)")
		return
	}
	printPowerResult(pm)
}

func cmdPower() {
	printSection("POWER & THERMAL (powermetrics)")
	pm, err := RunPowermetrics()
	if err != nil {
		fmt.Println("  (unavailable)")
		return
	}
	printPowerResult(pm)
}

func cmdDiag() {
	smc := openSMC()
	if smc == nil {
		os.Exit(1)
	}
	defer smc.Close()

	probeKeys := []string{
		"FNum",
		"F0Ac", "F0Mn", "F0Mx", "F0Tg",
		"Tp01", "Tp05", "Tp09", "Tp0D", "Tp0T",
		"Tp0h", "Tp0j", "Tp0l", "Tp0n", "Tp0t",
		"Tg0f", "Tg0j", "Tg0d",
		"Te05", "Te0L", "Te0P", "Te0S",
		"TB0T", "TB1T",
		"TaLP", "TaRF",
		"TW0P", "Ts0D",
		"TCAL", "TPCD",
	}

	printSection("SMC KEY DIAGNOSTICS")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  KEY\tTYPE\tRAW BYTES (hex)\tDECODED")
	fmt.Fprintln(w, "  ---\t----\t---------------\t-------")
	for _, k := range probeKeys {
		raw, typ, ok := smc.ReadRaw(k)
		if !ok {
			continue
		}
		hexStr := fmt.Sprintf("% x", raw)
		decoded := decodeRaw(typ, raw)
		fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", k, typ, hexStr, decoded)
	}
	w.Flush()
}

func decodeRaw(typ string, b []byte) string {
	switch typ {
	case "sp78":
		if len(b) >= 2 {
			raw := int16(b[0])<<8 | int16(b[1])
			return fmt.Sprintf("%.2f °C", float64(raw)/256.0)
		}
	case "flt ":
		if len(b) >= 4 {
			bits := uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
			v := *(*float32)(unsafe.Pointer(&bits))
			return fmt.Sprintf("%.2f", v)
		}
	case "fpe2":
		if len(b) >= 2 {
			rpm := float64((int(b[0])<<6 | int(b[1])>>2))
			return fmt.Sprintf("%.0f RPM", rpm)
		}
	case "ui8 ":
		if len(b) >= 1 {
			return fmt.Sprintf("%d", b[0])
		}
	case "ui16":
		if len(b) >= 2 {
			return fmt.Sprintf("%d", uint16(b[0])<<8|uint16(b[1]))
		}
	case "ui32":
		if len(b) >= 4 {
			return fmt.Sprintf("%d", uint32(b[0])<<24|uint32(b[1])<<16|uint32(b[2])<<8|uint32(b[3]))
		}
	}
	return "(unknown)"
}

func printPowerResult(pm *PowermetricsResult) {
	pressure := pm.ThermalPressure
	if pressure == "" {
		pressure = "unknown"
	}
	fmt.Printf("  Thermal Pressure : %s\n", pressure)
	fmt.Printf("  CPU Energy       : %.2f W\n", pm.CPUEnergyW/1000)
	fmt.Printf("  GPU Energy       : %.2f W\n", pm.GPUEnergyW/1000)

	if len(pm.Fans) > 0 {
		fmt.Println()
		fmt.Println("  Fans:")
		for _, f := range pm.Fans {
			fmt.Printf("    %-12s : %.0f RPM\n", f.Name, f.RPM)
		}
	}

	if len(pm.CPUClusters) > 0 {
		fmt.Println()
		fmt.Println("  CPU Clusters:")
		w := tabwriter.NewWriter(os.Stdout, 4, 0, 2, ' ', 0)
		for _, c := range pm.CPUClusters {
			freqMHz := c.Freq / 1e6
			fmt.Fprintf(w, "    %-16s\tFreq: %6.0f MHz\tActive: %.1f%%\n",
				c.Name, freqMHz, c.Active*100)
		}
		w.Flush()
	}
}

func printSection(title string) {
	line := strings.Repeat("─", len(title)+4)
	fmt.Printf("┌%s┐\n│  %s  │\n└%s┘\n", line, title, line)
}

func tempBar(c float64) string {
	const maxTemp = 100.0
	const barLen = 10
	filled := int(c / maxTemp * barLen)
	if filled > barLen {
		filled = barLen
	}
	if filled < 0 {
		filled = 0
	}

	var color string
	switch {
	case c < 50:
		color = "\033[32m"
	case c < 80:
		color = "\033[33m"
	default:
		color = "\033[31m"
	}
	reset := "\033[0m"

	bar := strings.Repeat("█", filled) + strings.Repeat("░", barLen-filled)
	return fmt.Sprintf("%s[%s]%s", color, bar, reset)
}
