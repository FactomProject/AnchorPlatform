package fees

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetETHFees(t *testing.T) {

	var err error

	jsonResp := []byte(`{"fast":1030,"fastest":1160,"safeLow":10,"average":960,"block_time":11.340909090909092,"blockNum":12198865,"speed":0.9983437183526395,"safeLowWait":8.8,"avgWait":0.7,"fastWait":0.4,"fastestWait":0.4,"gasPriceRange":{"4":189,"6":189,"8":189,"10":8.8,"20":8.8,"30":8.8,"40":8.8,"50":8.8,"60":8.8,"70":8.8,"80":8.8,"90":8.8,"100":8.2,"110":8.2,"120":8.2,"130":8.2,"140":8.2,"150":8.2,"160":8.2,"170":8.2,"180":8.2,"190":8.2,"200":8.2,"220":8.2,"240":8.2,"260":8.2,"280":8.2,"300":8.2,"320":7.7,"340":7.7,"360":7.7,"380":7.7,"400":7.7,"420":7.7,"440":7.7,"460":7.7,"480":7.7,"500":7.7,"520":7.7,"540":7.7,"560":7.7,"580":7.7,"600":7.7,"620":7.7,"640":7.7,"660":7.7,"680":7.7,"700":7.7,"720":7.7,"740":7.7,"760":7.7,"780":7.7,"800":7.7,"820":7.7,"840":7.7,"860":7.7,"880":7.7,"900":7.2,"920":7.2,"940":4,"960":0.7,"980":0.5,"1000":0.4,"1020":0.4,"1030":0.4,"1040":0.4,"1060":0.4,"1080":0.4,"1100":0.4,"1120":0.4,"1140":0.4,"1160":0.4}}`)
	fees := &ETHFees{}

	err = json.Unmarshal(jsonResp, fees)

	assert.NoError(t, err)
	assert.Equal(t, 1030, fees.Fast)
	assert.Equal(t, 1160, fees.Fastest)
	assert.Equal(t, 10, fees.SafeLow)
	assert.Equal(t, 960, fees.Average)

}

func TestGetBTCFees(t *testing.T) {

	var err error

	jsonResp := []byte(`{"fastestFee":100,"halfHourFee":91,"hourFee":82,"minimumFee":4}`)
	fees := &BTCFees{}

	err = json.Unmarshal(jsonResp, fees)

	assert.NoError(t, err)
	assert.Equal(t, 100, fees.FastestFee)
	assert.Equal(t, 91, fees.HalfHourFee)
	assert.Equal(t, 82, fees.HourFee)
	assert.Equal(t, 4, fees.MinimumFee)

}
