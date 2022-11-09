package main

var stateString = "0\n" + // protocol version
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
	"0x600023711f\n" + // level get: PREAMP|ATT|VOX|AF|RF|NR|RFPOWER|MICGAIN|KEYSPD|COMP|AGC|VOXGAIN|MONITOR_GAIN|NB (TODO: use metering protocol to add SWR|ALC|RFPOWER_METER|COMP_METER)
	"0x600023711f\n" + // level set: PREAMP|ATT|VOX|AF|RF|NR|RFPOWER|MICGAIN|KEYSPD|COMP|AGC|VOXGAIN|MONITOR_GAIN|NB
	"0\n" + // parm get: none
	"0\n" // parm set: none

func init() {
	hamlib.AddHandler(
		names{{`\dump_state`}},
		NewHandler(
			func(_ *Conn, _ []string) (string, error) {
				return stateString, nil
			},
			Args(0),
		),
	)
}
