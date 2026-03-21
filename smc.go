package main

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <stdint.h>
#include <stdio.h>
#include <string.h>
#include <IOKit/IOKitLib.h>

#define SMC_CMD_READ_BYTES    5
#define SMC_CMD_READ_KEYINFO  9
#define KERNEL_INDEX_SMC      2

#define SMC_KEY_FAN_NUM       "FNum"
#define SMC_KEY_FAN_SPEED     "F%dAc"
#define SMC_KEY_FAN_MIN       "F%dMn"
#define SMC_KEY_FAN_MAX       "F%dMx"

typedef struct {
    char          major;
    char          minor;
    char          build;
    char          reserved[1];
    uint16_t      release;
} SMCKeyData_vers_t;

typedef struct {
    uint16_t      version;
    uint16_t      length;
    uint32_t      cpuPLimit;
    uint32_t      gpuPLimit;
    uint32_t      memPLimit;
} SMCKeyData_pLimitData_t;

typedef struct {
    uint32_t      dataSize;
    uint32_t      dataType;
    char          dataAttributes;
} SMCKeyData_keyInfo_t;

typedef struct {
    uint32_t                  key;
    SMCKeyData_vers_t         vers;
    SMCKeyData_pLimitData_t   pLimitData;
    SMCKeyData_keyInfo_t      keyInfo;
    char                      result;
    char                      status;
    char                      data8;
    uint32_t                  data32;
    uint8_t                   bytes[32];
} SMCKeyData_t;

static io_connect_t g_conn = 0;

static uint32_t str_to_key(const char *str) {
    uint32_t key = 0;
    for (int i = 0; i < 4; i++) {
        key = (key << 8) | (uint8_t)str[i];
    }
    return key;
}

static kern_return_t smc_call(SMCKeyData_t *in, SMCKeyData_t *out) {
    size_t structureInputSize  = sizeof(SMCKeyData_t);
    size_t structureOutputSize = sizeof(SMCKeyData_t);
    return IOConnectCallStructMethod(
        g_conn,
        KERNEL_INDEX_SMC,
        in,  structureInputSize,
        out, &structureOutputSize
    );
}

int smc_open() {
    io_service_t service = IOServiceGetMatchingService(
        kIOMainPortDefault,
        IOServiceMatching("AppleSMC")
    );
    if (service == 0) return -1;
    kern_return_t result = IOServiceOpen(service, mach_task_self(), 0, &g_conn);
    IOObjectRelease(service);
    return (result == kIOReturnSuccess) ? 0 : -1;
}

void smc_close() {
    if (g_conn) {
        IOServiceClose(g_conn);
        g_conn = 0;
    }
}

int smc_read_key_info(const char *key, uint32_t *dataSize, uint32_t *dataType) {
    SMCKeyData_t in  = {0};
    SMCKeyData_t out = {0};
    in.key  = str_to_key(key);
    in.data8 = SMC_CMD_READ_KEYINFO;
    kern_return_t r = smc_call(&in, &out);
    if (r != kIOReturnSuccess || out.result != 0) return -1;
    *dataSize = out.keyInfo.dataSize;
    *dataType = out.keyInfo.dataType;
    return 0;
}

int smc_read_key(const char *key, uint8_t *bytes, uint32_t *dataSize) {
    uint32_t dataType = 0;
    if (smc_read_key_info(key, dataSize, &dataType) != 0) return -1;

    SMCKeyData_t in  = {0};
    SMCKeyData_t out = {0};
    in.key             = str_to_key(key);
    in.keyInfo.dataSize = *dataSize;
    in.data8           = SMC_CMD_READ_BYTES;
    kern_return_t r = smc_call(&in, &out);
    if (r != kIOReturnSuccess || out.result != 0) return -1;
    memcpy(bytes, out.bytes, *dataSize);
    return 0;
}

// Read number of fans
int smc_fan_count() {
    uint8_t  bytes[32] = {0};
    uint32_t size      = 0;
    if (smc_read_key(SMC_KEY_FAN_NUM, bytes, &size) != 0) return 0;
    return (int)bytes[0];
}

// Read fan actual RPM — supports both flt (Apple Silicon) and fpe2 (Intel) formats.
float smc_fan_rpm(int idx, int cmd_offset) {
    char key[5] = {0};
    if (cmd_offset == 0)      snprintf(key, 5, "F%dAc", idx);
    else if (cmd_offset == 1) snprintf(key, 5, "F%dMn", idx);
    else if (cmd_offset == 2) snprintf(key, 5, "F%dMx", idx);
    else if (cmd_offset == 3) snprintf(key, 5, "F%dTg", idx);

    uint32_t dataSize = 0, dataType = 0;
    if (smc_read_key_info(key, &dataSize, &dataType) != 0) return -1.0f;

    uint8_t bytes[32] = {0};
    if (smc_read_key(key, bytes, &dataSize) != 0) return -1.0f;

    // flt = 0x666c7420 (IEEE 754 float32, little-endian on Apple Silicon)
    if (dataType == 0x666c7420 && dataSize >= 4) {
        float v;
        memcpy(&v, bytes, 4);
        return v;
    }
    // fpe2 = 0x66706532 (unsigned fixed-point, big-endian, legacy Intel Macs)
    if (dataType == 0x66706532 && dataSize >= 2) {
        return (float)((bytes[0] << 6) | (bytes[1] >> 2));
    }
    return -1.0f;
}

// smc_key_datatype returns the 4-byte data type code of an SMC key as a uint32,
// or 0 if the key does not exist.
uint32_t smc_key_datatype(const char *key) {
    uint32_t dataSize = 0;
    uint32_t dataType = 0;
    if (smc_read_key_info(key, &dataSize, &dataType) != 0) return 0;
    return dataType;
}

// smc_read_raw fills bytes with up to 32 raw bytes for a key.
// Returns actual data size, or -1 on error.
int smc_read_raw(const char *key, uint8_t *bytes, uint32_t *dtype) {
    uint32_t dataSize = 0;
    if (smc_read_key_info(key, &dataSize, dtype) != 0) return -1;
    if (smc_read_key(key, bytes, &dataSize) != 0) return -1;
    return (int)dataSize;
}

// Read temperature key - supports sp78 and flt (float32) types.
// Returns -300.0 if the key doesn't exist or is an unrecognised type.
float smc_temp(const char *key) {
    uint32_t dataSize = 0;
    uint32_t dataType = 0;
    if (smc_read_key_info(key, &dataSize, &dataType) != 0) return -300.0f;

    uint8_t bytes[32] = {0};
    if (smc_read_key(key, bytes, &dataSize) != 0) return -300.0f;

    // sp78 = 0x73703738  (signed fixed-point 7.8)
    if (dataType == 0x73703738) {
        int16_t raw = (int16_t)((bytes[0] << 8) | bytes[1]);
        return (float)raw / 256.0f;
    }
    // flt  = 0x666c7420  (IEEE 754 float32)
    if (dataType == 0x666c7420 && dataSize >= 4) {
        float v;
        memcpy(&v, bytes, 4);
        return v;
    }
    // sp5a = 0x73703561  (signed fixed-point 5.10, used on some M-series chips)
    if (dataType == 0x73703561) {
        int16_t raw = (int16_t)((bytes[0] << 8) | bytes[1]);
        return (float)raw / 1024.0f;
    }
    return -300.0f;
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// SMC manages the connection to Apple's System Management Controller.
type SMC struct {
	opened bool
}

// Open initialises the SMC connection.
func (s *SMC) Open() error {
	if rc := C.smc_open(); rc != 0 {
		return fmt.Errorf("failed to open SMC connection (IOKit error %d)", rc)
	}
	s.opened = true
	return nil
}

// Close releases the SMC connection.
func (s *SMC) Close() {
	if s.opened {
		C.smc_close()
		s.opened = false
	}
}

// FanCount returns the number of fans reported by SMC.
func (s *SMC) FanCount() int {
	return int(C.smc_fan_count())
}

// FanRPM returns actual, min, max, target RPM for fan idx.
func (s *SMC) FanRPM(idx int) (actual, min, max, target float64) {
	actual = float64(C.smc_fan_rpm(C.int(idx), 0))
	min = float64(C.smc_fan_rpm(C.int(idx), 1))
	max = float64(C.smc_fan_rpm(C.int(idx), 2))
	target = float64(C.smc_fan_rpm(C.int(idx), 3))
	return
}

// Temp reads a temperature key in degrees Celsius.
// Returns (value, true) on success, or (0, false) if the key is unavailable.
func (s *SMC) Temp(key string) (float64, bool) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))
	v := float64(C.smc_temp(ckey))
	if v < -200 {
		return 0, false
	}
	return v, true
}

// TemperatureSensors returns all known temperature sensor readings.
func (s *SMC) TemperatureSensors() []TempSensor {
	// Apple Silicon (M1-M4) SMC temperature keys
	knownKeys := []struct {
		key  string
		name string
	}{
		// CPU cores
		{"Tp01", "CPU Core 1"},
		{"Tp05", "CPU Core 2"},
		{"Tp09", "CPU Core 3 (Perf)"},
		{"Tp0D", "CPU Core 4 (Perf)"},
		{"Tp0b", "CPU Core 5"},
		{"Tp0f", "CPU Core 6"},
		{"Tp0h", "CPU Core 7"},
		{"Tp0j", "CPU Core 8"},
		{"Tp0l", "CPU Core 9"},
		{"Tp0n", "CPU Core 10"},
		{"Tp0p", "CPU Core 11"},
		{"Tp0r", "CPU Core 12"},
		{"Tp0t", "CPU Core 13"},
		{"Tp0T", "CPU Package"},
		{"TCAL", "CPU Calibrated"},
		// GPU
		{"Tg0f", "GPU Core 1"},
		{"Tg0j", "GPU Core 2"},
		{"Tg0d", "GPU Core 3"},
		{"Tg0b", "GPU Core 4"},
		{"Tg0h", "GPU Core 5"},
		{"Tg0l", "GPU Core 6"},
		{"Tg0p", "GPU Core 7"},
		{"Tg0t", "GPU Core 8"},
		// SoC / ANE
		{"Te05", "SoC Energy"},
		{"Te0L", "SoC 2"},
		{"Te0P", "SoC 3"},
		{"Te0S", "SoC 4"},
		// Memory
		{"TM0P", "DRAM"},
		{"TM1P", "DRAM 2"},
		// Battery
		{"TB0T", "Battery 1"},
		{"TB1T", "Battery 2"},
		{"TB2T", "Battery 3"},
		{"TB3T", "Battery 4"},
		// Ambient / PCB
		{"TaLP", "Ambient Left"},
		{"TaRF", "Ambient Right"},
		{"TW0P", "WiFi"},
		{"TPCD", "PCH"},
		{"Ts0D", "NAND"},
		{"Ts0P", "NAND 2"},
	}

	var results []TempSensor
	for _, k := range knownKeys {
		v, ok := s.Temp(k.key)
		// Filter out: unavailable keys, physically impossible temps, and near-zero placeholders.
		// Apple Silicon returns ~0-2°C for power-gated CPU cores; ambient is always > 10°C.
		if ok && v > 10.0 && v < 120.0 {
			results = append(results, TempSensor{Key: k.key, Name: k.name, Celsius: v})
		}
	}
	return results
}

// TempSensor holds a single temperature reading.
type TempSensor struct {
	Key     string
	Name    string
	Celsius float64
}

// KeyDataType returns the raw 4-byte type code for a key as a printable string (e.g. "sp78").
func (s *SMC) KeyDataType(key string) string {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))
	dt := uint32(C.smc_key_datatype(ckey))
	if dt == 0 {
		return ""
	}
	b := []byte{byte(dt >> 24), byte(dt >> 16), byte(dt >> 8), byte(dt)}
	return string(b)
}

// ReadRaw reads raw bytes for a key, returns (data, typeStr, ok).
func (s *SMC) ReadRaw(key string) ([]byte, string, bool) {
	ckey := C.CString(key)
	defer C.free(unsafe.Pointer(ckey))
	var dtype C.uint32_t
	buf := make([]byte, 32)
	n := int(C.smc_read_raw(ckey, (*C.uint8_t)(unsafe.Pointer(&buf[0])), &dtype))
	if n < 0 {
		return nil, "", false
	}
	dt := uint32(dtype)
	b := []byte{byte(dt >> 24), byte(dt >> 16), byte(dt >> 8), byte(dt)}
	return buf[:n], string(b), true
}
