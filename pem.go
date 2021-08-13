package ydb

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
)

// content of https://storage.yandexcloud.net/cloud-certs/CA.pem
var ydbPEM = []byte(`
-----BEGIN CERTIFICATE-----
MIIE3TCCAsWgAwIBAgIKPxb5sAAAAAAAFzANBgkqhkiG9w0BAQ0FADAfMR0wGwYD
VQQDExRZYW5kZXhJbnRlcm5hbFJvb3RDQTAeFw0xNzA2MjAxNjQ0MzdaFw0yNzA2
MjAxNjU0MzdaMFUxEjAQBgoJkiaJk/IsZAEZFgJydTEWMBQGCgmSJomT8ixkARkW
BnlhbmRleDESMBAGCgmSJomT8ixkARkWAmxkMRMwEQYDVQQDEwpZYW5kZXhDTENB
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqgNnjk0JKPcbsk1+KG2t
eM1AfMnEe5RkAJuBBuwVV49snhcvO1jhKBx/pCnjr6biICc1/oAFDVgU8yVYYPwp
WZ2vH3ZtscjJ/RAT/NS9OKKG7kKknhFhVYxua5xhoIQmm6usBNYYiTcWoFm1eHC8
I9oddOLSscZYbh3unVRvt+3V+drVmUx9oSUKpqMgfysiv1MN6zB3vq9TFkbhz53E
k0tEcV+W2NnDaeFhLKy284FDKLvOdTDj1EDsSAihxl7sNEKpupNuhgyy2siOqUb+
d5mO/CRfaAKGg3E6hDM3pEi48E506dJdjPXWfHKSvuguMLRlb2RWdVocRZuyWxOh
0QIDAQABo4HkMIHhMBAGCSsGAQQBgjcVAQQDAgEAMB0GA1UdDgQWBBRMU5uItjx+
TOicX1+ovC1Xq2PSnzAZBgkrBgEEAYI3FAIEDB4KAFMAdQBiAEMAQTALBgNVHQ8E
BAMCAYYwDwYDVR0TAQH/BAUwAwEB/zAfBgNVHSMEGDAWgBSrucX/oe/mUx0zOSKE
0XbUN04tajBUBgNVHR8ETTBLMEmgR6BFhkNodHRwOi8vY3Jscy55YW5kZXgucnUv
WWFuZGV4SW50ZXJuYWxSb290Q0EvWWFuZGV4SW50ZXJuYWxSb290Q0EuY3JsMA0G
CSqGSIb3DQEBDQUAA4ICAQAsR5Lb4Pv2FD0Kk+4oc1GEOnehxKLsQtdV81nrU+IV
l9pr2oNMdi8lwIolvHZRllLM4Ba5AcRH6YJ5fe7AjKm+5EdSkhqVWo2UOllRCbtS
wmL50+erOAkxstSlRkO6b8x1L0MOBKv54E5YcQ/Wwt27ldSb6RkEmJBGvmxObAaf
5zc51pqSqao9tnldYaCblEQ/Zmy43FliIpa2eUJoh8DqK8bVo2gcI3wbQ32tWs9u
wvKk8fo4lAdhCwhv+QHuqau1VAY9hPU106bsFIDUmijTMxjAobKBi6CkIX6EbNHU
Jv4DzYVLlDd2y0CADdn2F6I70xpCBn5cquSGuvFbqZjQDmIHwb7WQSxadkiGRWfc
zVTnmiHjJONJJIpE2t+FOV3hc+8o98OzOtNaH2QQ9j6dnKvtIGKGFeNSDp0vXPOi
QhHiIyuB7eWx+g2whktQ74UCpGDSXYnEW3s8w5wezVWIEmouq7q4rCEkTNvJ7Ico
43AgUdPzAFS2zYktw1C+cbUALM8smvXbXrXOBzMmscjIhtXvLMrpPeh23VfdJfQB
0rN2BmRCLUE8JOV+o0k98XMm83oN+lGkL1l+hyoj3ok1uI3JrsWOcDyjOds3ptcN
KimJLm27ndjcxDNo/iA6gefMJuCxFRaqI+eF4P0jSkMgnnQqZkvLGFuHCw8eRDhm
bw==
-----END CERTIFICATE-----
-----BEGIN CERTIFICATE-----
MIIFGTCCAwGgAwIBAgIQJMM7ZIy2SYxCBgK7WcFwnjANBgkqhkiG9w0BAQ0FADAf
MR0wGwYDVQQDExRZYW5kZXhJbnRlcm5hbFJvb3RDQTAeFw0xMzAyMTExMzQxNDNa
Fw0zMzAyMTExMzUxNDJaMB8xHTAbBgNVBAMTFFlhbmRleEludGVybmFsUm9vdENB
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAgb4xoQjBQ7oEFk8EHVGy
1pDEmPWw0Wgw5nX9RM7LL2xQWyUuEq+Lf9Dgh+O725aZ9+SO2oEs47DHHt81/fne
5N6xOftRrCpy8hGtUR/A3bvjnQgjs+zdXvcO9cTuuzzPTFSts/iZATZsAruiepMx
SGj9S1fGwvYws/yiXWNoNBz4Tu1Tlp0g+5fp/ADjnxc6DqNk6w01mJRDbx+6rlBO
aIH2tQmJXDVoFdrhmBK9qOfjxWlIYGy83TnrvdXwi5mKTMtpEREMgyNLX75UjpvO
NkZgBvEXPQq+g91wBGsWIE2sYlguXiBniQgAJOyRuSdTxcJoG8tZkLDPRi5RouWY
gxXr13edn1TRDGco2hkdtSUBlajBMSvAq+H0hkslzWD/R+BXkn9dh0/DFnxVt4XU
5JbFyd/sKV/rF4Vygfw9ssh1ZIWdqkfZ2QXOZ2gH4AEeoN/9vEfUPwqPVzL0XEZK
r4s2WjU9mE5tHrVsQOZ80wnvYHYi2JHbl0hr5ghs4RIyJwx6LEEnj2tzMFec4f7o
dQeSsZpgRJmpvpAfRTxhIRjZBrKxnMytedAkUPguBQwjVCn7+EaKiJfpu42JG8Mm
+/dHi+Q9Tc+0tX5pKOIpQMlMxMHw8MfPmUjC3AAd9lsmCtuybYoeN2IRdbzzchJ8
l1ZuoI3gH7pcIeElfVSqSBkCAwEAAaNRME8wCwYDVR0PBAQDAgGGMA8GA1UdEwEB
/wQFMAMBAf8wHQYDVR0OBBYEFKu5xf+h7+ZTHTM5IoTRdtQ3Ti1qMBAGCSsGAQQB
gjcVAQQDAgEAMA0GCSqGSIb3DQEBDQUAA4ICAQAVpyJ1qLjqRLC34F1UXkC3vxpO
nV6WgzpzA+DUNog4Y6RhTnh0Bsir+I+FTl0zFCm7JpT/3NP9VjfEitMkHehmHhQK
c7cIBZSF62K477OTvLz+9ku2O/bGTtYv9fAvR4BmzFfyPDoAKOjJSghD1p/7El+1
eSjvcUBzLnBUtxO/iYXRNo7B3+1qo4F5Hz7rPRLI0UWW/0UAfVCO2fFtyF6C1iEY
/q0Ldbf3YIaMkf2WgGhnX9yH/8OiIij2r0LVNHS811apyycjep8y/NkG4q1Z9jEi
VEX3P6NEL8dWtXQlvlNGMcfDT3lmB+tS32CPEUwce/Ble646rukbERRwFfxXojpf
C6ium+LtJc7qnK6ygnYF4D6mz4H+3WaxJd1S1hGQxOb/3WVw63tZFnN62F6/nc5g
6T44Yb7ND6y3nVcygLpbQsws6HsjX65CoSjrrPn0YhKxNBscF7M7tLTW/5LK9uhk
yjRCkJ0YagpeLxfV1l1ZJZaTPZvY9+ylHnWHhzlq0FzcrooSSsp4i44DB2K7O2ID
87leymZkKUY6PMDa4GkDJx0dG4UXDhRETMf+NkYgtLJ+UIzMNskwVDcxO4kVL+Hi
Pj78bnC5yCw8P5YylR45LdxLzLO68unoXOyFz1etGXzszw8lJI9LNubYxk77mK8H
LpuQKbSbIERsmR+QqQ==
-----END CERTIFICATE-----
`)

func AppendCertsFromFile(certPool *x509.CertPool, caFile string) error {
	pem, err := ioutil.ReadFile(caFile)
	if err != nil {
		return err
	}
	if ok := certPool.AppendCertsFromPEM(pem); !ok {
		return fmt.Errorf("certificates file '%s' can not be append: %w", caFile, err)
	}
	return nil

}

func WithYdbCA(certPool *x509.CertPool) (_ *x509.CertPool, err error) {
	if certPool == nil {
		certPool, err = x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
	}
	// append user defined certs
	if caFile, ok := os.LookupEnv("YDB_SSL_ROOT_CERTIFICATES_FILE"); ok {
		if err := AppendCertsFromFile(certPool, caFile); err != nil {
			return nil, err
		}
	}
	// append internal YDB certs
	if ok := certPool.AppendCertsFromPEM(ydbPEM); !ok {
		return nil, fmt.Errorf("internal YDB certs can not be append")
	}
	return certPool, nil
}
