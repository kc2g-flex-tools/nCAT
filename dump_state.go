package main

import (
	"context"
	"fmt"
)

var stateString = "1\n" + // protocol version
	"2\n" + // hamlib model
	"2\n" + // region
	"30000.000000 54000000.000000 0xe2f -1 -1 0x1 0x0\n" + // RX: 30kHz - 54MHz, AM|CW|USB|LSB|FM|AMS|PKTUSB|PKTLSB
	"0 0 0 0 0 0 0\n" + // end of RX range list
	"100000.000000 54000000.000000 0xe2f 1000 100000 0x1 0x0\n" + // TX: 100kHz - 54MHz, 1-100 watts, AM|CW|USB|LSB|FM|AMS|PKTUSB|PKTLSB
	"0 0 0 0 0 0 0\n" + // end of TX range list
	"0xe2f 1\n" +
	"0xe2f 0\n" +
	"0 0\n" + // end of tuning steps
	"0x02 500\n" + // CW normal
	"0x02 200\n" + // CW narrow
	"0x02 2000\n" + // CW wide
	"0x02 50\n" + // CW min
	"0x02 10000\n" + // CW max
	"0x02 0\n" + // allow arbitrary
	"0x221 6000\n" + // AM|FM|AMS normal
	"0x221 3000\n" + // AM|FM|AMS narrow
	"0x221 10000\n" + // AM|FM|AMS wide
	"0x221 100\n" + // AM|FM|AMS min
	"0x221 20000\n" + // AM|FM|AMS max
	"0x221 0\n" + // allow arbitrary
	"0x0c 2700\n" + // SSB normal
	"0x0c 1400\n" + // SSB narrow
	"0x0c 3900\n" + // SSB wide
	"0x0c 50\n" + // SSB min
	"0x0c 10000\n" + // SSB max
	"0x0c 0\n" + // allow arbitrary
	"0xc00 3000\n" + // digi normal
	"0xc00 1500\n" + // digi narrow
	"0xc00 4000\n" + // digi wide
	"0xc00 50\n" + // digi min
	"0xc00 10000\n" + //digi max
	"0xc00 0\n" + // allow arbitrary
	"0 0\n" + // end of filter widths
	"0\n" + // max rit
	"0\n" + // max xit
	"0\n" + // max if_shift
	"0\n" + // no announce capabilities
	"0 8 16 24 32\n" + // preamp
	"0 8\n" + // attenuator
	"0x48400833be\n" + // func get: NB|COMP|VOX|TONE|TSQL|FBKIN|ANF|NR|MON|MN|REV|TUNER|ANL|DIVERSITY
	"0x48400833be\n" + // func set: NB|COMP|VOX|TONE|TSQL|FBKIN|ANF|NR|MON|MN|REV|TUNER|ANL|DIVERSITY
	"0x%x\n" + // level get: TBD at runtime
	"0x600023711f\n" + // level set: PREAMP|ATT|VOX|AF|RF|NR|RFPOWER|MICGAIN|KEYSPD|COMP|AGC|VOXGAIN|MONITOR_GAIN|NB
	"0\n" + // parm get: none
	"0\n" // parm set: none

var protocol1StateString = "vfo_ops=0x0\n" +
	"targetable_vfo=0\n" +
	"has_set_vfo=0\n" +
	"has_get_vfo=1\n" +
	"has_set_conf=0\n" +
	"has_get_conf=0\n" +
	"has_set_freq=1\n" +
	"has_get_freq=1\n" +
	"done\n"

func init() {
	hamlib.AddHandler(
		NewHandler(
			"dump_state", "",
			func(ctx context.Context, _ []string) (string, error) {
				conn := getConn(ctx)
				var levelGetCaps uint64
				if !cfg.Metering {
					// PREAMP|ATT|VOX|AF|RF|NR|RFPOWER|MICGAIN|KEYSPD|COMP|AGC|VOXGAIN|MONITOR_GAIN|NB
					levelGetCaps = 0x600023711f
				} else {
					// As above plus SWR|ALC|STRENGTH|RFPOWER_METER|COMP_METER|VD_METER|RFPOWER_METER_WATTS|TEMP_METER
					levelGetCaps = 0x100e77023711f
				}
				var ret = fmt.Sprintf(stateString, levelGetCaps)
				if conn.chkVFOexecuted { // match hamlib behavior here
					ret += protocol1StateString
				}
				return ret, nil
			},
			Args(0),
		),
	)
}
