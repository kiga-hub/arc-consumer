package ws

// AudioEngine - 音频算法返回结构
type AudioEngine struct {
	Time int64 `json:"time"`

	//statistics
	Mean    *float32 `json:"mean,omitempty"`
	Varance *float32 `json:"varance,omitempty"`
	Skew    *float32 `json:"skew,omitempty"`
	Kurt    *float32 `json:"kurt,omitempty"`
	Std     *float32 `json:"std,omitempty"`
	Rms     *float32 `json:"rms,omitempty"`
	Minv    *float32 `json:"minv,omitempty"`
	Maxv    *float32 `json:"maxv,omitempty"`
	Peak    *float32 `json:"peak,omitempty"`
	Ppv     *float32 `json:"ppv,omitempty"`

	//dimensionless indicators
	IndMargin   *float32 `json:"ind_margin,omitempty"`
	IndWaveform *float32 `json:"ind_waveform,omitempty"`
	IndImpulse  *float32 `json:"ind_impulse,omitempty"`
	IndPeak     *float32 `json:"ind_peak,omitempty"`
	IndKurt     *float32 `json:"ind_kurt,omitempty"`
	//raw
	Raw *float32 `json:"raw,omitempty"`

	// 音频强度
	DB *float32 `json:"db,omitempty"`

	// Spark 电火花引擎结果
	Spark *float64 `json:"spark,omitempty"`
}

// DeepCopy -
func (e *AudioEngine) DeepCopy(d *AudioEngine) {
	d.Time = e.Time
	if e.Spark != nil {
		d.NewSpark()
		*d.Spark = *e.Spark
	}
	if e.Raw != nil {
		d.NewVATD()

		*d.Mean = *e.Mean
		*d.Varance = *e.Varance
		*d.Skew = *e.Skew
		*d.Kurt = *e.Kurt
		*d.Std = *e.Std
		*d.Rms = *e.Rms
		*d.Minv = *e.Minv
		*d.Maxv = *e.Maxv
		*d.Peak = *e.Peak
		*d.Ppv = *e.Ppv

		*d.IndMargin = *e.IndMargin
		*d.IndWaveform = *e.IndWaveform
		*d.IndImpulse = *e.IndImpulse
		*d.IndPeak = *e.IndPeak
		*d.IndKurt = *e.IndKurt

		*d.Raw = *e.Raw
		*d.DB = *e.DB
	}
}

// AudioEngineData -
type AudioEngineData struct {
	CollectorID string        `json:"collectorid"`
	Data        []AudioEngine `json:"data"`
}

// AudioEngineDataResponse is the response for getting vibrates
type AudioEngineDataResponse struct {
	Code int               `json:"code,omitempty"`
	Msg  string            `json:"msg,omitempty"`
	Data []AudioEngineData `json:"data,omitempty"`
}

// DeepCopy -
func (d *AudioEngineData) DeepCopy() *AudioEngineData {
	tmp := new(AudioEngineData)
	tmp.Data = make([]AudioEngine, len(d.Data))
	tmp.CollectorID = d.CollectorID
	for i, d := range d.Data {
		d.DeepCopy(&tmp.Data[i])
	}
	return tmp
}

// NewSpark -
func (e *AudioEngine) NewSpark() {
	if e.Spark != nil {
		return
	}
	e.Spark = new(float64)
}

// ClearSpark -
func (e *AudioEngine) ClearSpark() {
	if e.Spark == nil {
		return
	}
	*e.Spark = 0
}

// NewVATD -
func (e *AudioEngine) NewVATD() {
	if e.Raw != nil {
		return
	}
	e.Mean = new(float32)
	e.Varance = new(float32)
	e.Skew = new(float32)
	e.Kurt = new(float32)
	e.Std = new(float32)
	e.Rms = new(float32)
	e.Minv = new(float32)
	e.Maxv = new(float32)
	e.Peak = new(float32)
	e.Ppv = new(float32)

	e.IndMargin = new(float32)
	e.IndWaveform = new(float32)
	e.IndImpulse = new(float32)
	e.IndPeak = new(float32)
	e.IndKurt = new(float32)

	e.Raw = new(float32)
	e.DB = new(float32)
}

// ClearVATD -
func (e *AudioEngine) ClearVATD() {
	if e.Raw == nil {
		return
	}
	*e.Mean = 0
	*e.Varance = 0
	*e.Skew = 0
	*e.Kurt = 0
	*e.Std = 0
	*e.Rms = 0
	*e.Minv = 0
	*e.Maxv = 0
	*e.Peak = 0
	*e.Ppv = 0

	*e.IndMargin = 0
	*e.IndWaveform = 0
	*e.IndImpulse = 0
	*e.IndPeak = 0
	*e.IndKurt = 0

	*e.Raw = 0
	*e.DB = 0
}
