package part5

import (
	"encoding/binary"
	"errors"
	"math"
	"time"

	"github.com/pascaldekloe/part5/info"
)

// A Monitor consumes information in monitor direction as listed in table 8
// from companion standard 101.
type Monitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	SinglePtMonitor[Orig, Com, Obj]
	SinglePtChangeMonitor[Orig, Com, Obj]
	DoublePtMonitor[Orig, Com, Obj]
	StepMonitor[Orig, Com, Obj]
	BitsMonitor[Orig, Com, Obj]
	NormMonitor[Orig, Com, Obj]
	ScaledMonitor[Orig, Com, Obj]
	FloatMonitor[Orig, Com, Obj]
	TotalsMonitor[Orig, Com, Obj]
	ProtectMonitor[Orig, Com, Obj]
	ProtectStartMonitor[Orig, Com, Obj]
	ProtectOutMonitor[Orig, Com, Obj]
	InitEndMonitor[Orig, Com, Obj]
}

// SinglePtMonitor consumes single points.
type SinglePtMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// SinglePt gets called for type identifier 1: M_SP_NA_1.
	SinglePt(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual)
	// SinglePtAtMinute gets called for type identifier 2: M_SP_TA_1.
	SinglePtAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, info.CP24Time2a)
	// SinglePtAtMoment gets called for type identifier 30: M_SP_TB_1.
	SinglePtAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, info.CP56Time2a)
}

type singlePtProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	listener   func(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, time.Time)
	timeZone   *time.Location
	timeLeeway time.Duration
}

// SinglePointProxy abstracts the variants from the SinglePointMonitor interface
// into one function. Time tags are interpretated within the time-zone argument.
// The time.Time argument is always zero for type M_SP_NA_1. Time is also zero
// for Invalid() info.CP24Time2a and info.CP56Time2a.
//
// Time tags are assumed to be recent. See the example of WithinHourBefore from
// info.CP24Time2a for an explaination of the leeway setting.
// https://pkg.go.dev/github.com/pascaldekloe/part5/info#example-CP24Time2a.WithinHourBefore
func SinglePtProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](listener func(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtQual, time.Time), zone *time.Location, leeway time.Duration) SinglePtMonitor[Orig, Com, Obj] {
	return singlePtProxy[Orig, Com, Obj]{listener, zone, leeway}
}

func (proxy singlePtProxy[Orig, Com, Obj]) SinglePt(u info.DataUnit[Orig, Com, Obj], addr Obj, pt info.SinglePtQual) {
	proxy.listener(u, addr, pt, time.Time{})
}

func (proxy singlePtProxy[Orig, Com, Obj]) SinglePtAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, pt info.SinglePtQual, tag info.CP24Time2a) {
	proxy.listener(u, addr, pt, tag.WithinHourBefore(time.Now().In(proxy.timeZone).Add(proxy.timeLeeway)))
}

func (proxy singlePtProxy[Orig, Com, Obj]) SinglePtAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, pt info.SinglePtQual, tag info.CP56Time2a) {
	proxy.listener(u, addr, pt, tag.Within20thCentury(proxy.timeZone))
}

// SinglePtChangeMonitor consumes single points with change detection.
type SinglePtChangeMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// SinglePtChangePack gets called for type identifier 20: M_PS_NA_1.
	SinglePtChangePack(info.DataUnit[Orig, Com, Obj], Obj, info.SinglePtChangePack, info.Qual)
}

// DoublePtMonitor consumes double points.
type DoublePtMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// DoublePt gets called for type identifier 3: M_DP_NA_1.
	DoublePt(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual)
	// DoublePtAtMinute gets called for type identifier 4: M_DP_TA_1.
	DoublePtAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, info.CP24Time2a)
	// DoublePtAtMoment gets called for type identifier 31: M_DP_TB_1.
	DoublePtAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, info.CP56Time2a)
}

type doublePtProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	listener   func(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, time.Time)
	timeZone   *time.Location
	timeLeeway time.Duration
}

// DoublePointProxy abstracts the variants from the DoublePointMonitor interface
// into one function. Time tags are interpretated within the time-zone argument.
// The time.Time argument is always zero for type M_DP_NA_1. Time is also zero
// for Invalid() info.CP24Time2a and info.CP56Time2a.
//
// Time tags are assumed to be recent. See the example of WithinHourBefore from
// info.CP24Time2a for an explaination of the leeway setting.
// https://pkg.go.dev/github.com/pascaldekloe/part5/info#example-CP24Time2a.WithinHourBefore
func DoublePtProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](listener func(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, time.Time), zone *time.Location, leeway time.Duration) DoublePtMonitor[Orig, Com, Obj] {
	return doublePtProxy[Orig, Com, Obj]{listener, zone, leeway}
}

func (proxy doublePtProxy[Orig, Com, Obj]) DoublePt(u info.DataUnit[Orig, Com, Obj], addr Obj, pt info.DoublePtQual) {
	proxy.listener(u, addr, pt, time.Time{})
}

func (proxy doublePtProxy[Orig, Com, Obj]) DoublePtAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, pt info.DoublePtQual, tag info.CP24Time2a) {
	proxy.listener(u, addr, pt, tag.WithinHourBefore(time.Now().In(proxy.timeZone).Add(proxy.timeLeeway)))
}

func (proxy doublePtProxy[Orig, Com, Obj]) DoublePtAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, pt info.DoublePtQual, tag info.CP56Time2a) {
	proxy.listener(u, addr, pt, tag.Within20thCentury(proxy.timeZone))
}

// StepMonitor consumes step positions.
type StepMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// Step gets called for type identifier 5: M_ST_NA_1.
	Step(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual)
	// StepAtMinute gets called for type identifier 6: M_ST_TA_1.
	StepAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, info.CP24Time2a)
	// StepAtMoment gets called for type identifier 32: M_ST_TB_1.
	StepAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, info.CP56Time2a)
}

type stepProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	listener   func(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, time.Time)
	timeZone   *time.Location
	timeLeeway time.Duration
}

// StepProxy abstracts the variants from the StepMonitor interface
// into one function. Time tags are interpretated within the time-zone argument.
// The time.Time argument is always zero for type M_ST_NA_1. Time is also zero
// for Invalid() info.CP24Time2a and info.CP56Time2a.
//
// Time tags are assumed to be recent. See the example of WithinHourBefore from
// info.CP24Time2a for an explaination of the leeway setting.
// https://pkg.go.dev/github.com/pascaldekloe/part5/info#example-CP24Time2a.WithinHourBefore
func StepProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](listener func(info.DataUnit[Orig, Com, Obj], Obj, info.StepQual, time.Time), zone *time.Location, leeway time.Duration) StepMonitor[Orig, Com, Obj] {
	return stepProxy[Orig, Com, Obj]{listener, zone, leeway}
}

func (proxy stepProxy[Orig, Com, Obj]) Step(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual) {
	proxy.listener(u, addr, p, time.Time{})
}

func (proxy stepProxy[Orig, Com, Obj]) StepAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, tag info.CP24Time2a) {
	proxy.listener(u, addr, p, tag.WithinHourBefore(time.Now().In(proxy.timeZone).Add(proxy.timeLeeway)))
}

func (proxy stepProxy[Orig, Com, Obj]) StepAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, p info.StepQual, tag info.CP56Time2a) {
	proxy.listener(u, addr, p, tag.Within20thCentury(proxy.timeZone))
}

// BitsMonitor consumes bitstrings.
type BitsMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// Bits gets called for type identifier 7: M_BO_NA_1.
	Bits(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual)
	// BitsAtMinute gets called for type identifier 8: M_BO_TA_1.
	BitsAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, info.CP24Time2a)
	// BitsAtMoment gets called for type identifier 33: M_BO_TB_1.
	BitsAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, info.CP56Time2a)
}

type bitsProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	listener   func(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, time.Time)
	timeZone   *time.Location
	timeLeeway time.Duration
}

// BitsProxy abstracts the variants from the BitsMonitor interface
// into one function. Time tags are interpretated within the time-zone argument.
// The time.Time argument is always zero for type M_BO_NA_1. Time is also zero
// for Invalid() info.CP24Time2a and info.CP56Time2a.
//
// Time tags are assumed to be recent. See the example of WithinHourBefore from
// info.CP24Time2a for an explaination of the leeway setting.
// https://pkg.go.dev/github.com/pascaldekloe/part5/info#example-CP24Time2a.WithinHourBefore
func BitsProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](listener func(info.DataUnit[Orig, Com, Obj], Obj, info.BitsQual, time.Time), zone *time.Location, leeway time.Duration) BitsMonitor[Orig, Com, Obj] {
	return bitsProxy[Orig, Com, Obj]{listener, zone, leeway}
}

func (proxy bitsProxy[Orig, Com, Obj]) Bits(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual) {
	proxy.listener(u, addr, b, time.Time{})
}

func (proxy bitsProxy[Orig, Com, Obj]) BitsAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual, tag info.CP24Time2a) {
	proxy.listener(u, addr, b, tag.WithinHourBefore(time.Now().In(proxy.timeZone).Add(proxy.timeLeeway)))
}

func (proxy bitsProxy[Orig, Com, Obj]) BitsAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, b info.BitsQual, tag info.CP56Time2a) {
	proxy.listener(u, addr, b, tag.Within20thCentury(proxy.timeZone))
}

// NormMonitor consumes normalized values.
type NormMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// NormUnqual gets called for type identifier 21: M_ME_ND_1.
	NormUnqual(info.DataUnit[Orig, Com, Obj], Obj, info.Norm)
	// Norm gets called for type identifier 9: M_ME_NA_1.
	Norm(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual)
	// NormAtMinute gets called for type identifier 10: M_ME_TA_1.
	NormAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, info.CP24Time2a)
	// NormAtMoment gets called for type identifier 34: M_ME_TD_1.
	NormAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, info.CP56Time2a)
}

type normProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	listener   func(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, time.Time)
	timeZone   *time.Location
	timeLeeway time.Duration
}

// NormProxy abstracts the variants from the NormMonitor interface into one
// function. Time tags are interpretated within the time-zone argument. The
// time.Time argument is always zero for type M_ME_ND_1 and M_ME_NA_1. Time
// is also zero for Invalid() info.CP24Time2a and info.CP56Time2a. The
// quality descriptor is always info.OK for type M_ME_ND_1.
//
// Time tags are assumed to be recent. See the example of WithinHourBefore from
// info.CP24Time2a for an explaination of the leeway setting.
// https://pkg.go.dev/github.com/pascaldekloe/part5/info#example-CP24Time2a.WithinHourBefore
func NormProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](listener func(info.DataUnit[Orig, Com, Obj], Obj, info.NormQual, time.Time), zone *time.Location, leeway time.Duration) NormMonitor[Orig, Com, Obj] {
	return normProxy[Orig, Com, Obj]{listener, zone, leeway}
}

func (proxy normProxy[Orig, Com, Obj]) NormUnqual(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.Norm) {
	proxy.listener(u, addr, info.NormQual{n[0], n[1], uint8(info.OK)}, time.Time{})
}

func (proxy normProxy[Orig, Com, Obj]) Norm(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual) {
	proxy.listener(u, addr, n, time.Time{})
}

func (proxy normProxy[Orig, Com, Obj]) NormAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, tag info.CP24Time2a) {
	proxy.listener(u, addr, n, tag.WithinHourBefore(time.Now().In(proxy.timeZone).Add(proxy.timeLeeway)))
}

func (proxy normProxy[Orig, Com, Obj]) NormAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, n info.NormQual, tag info.CP56Time2a) {
	proxy.listener(u, addr, n, tag.Within20thCentury(proxy.timeZone))
}

// ScaledMonitor consumes scaled values.
type ScaledMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// Scaled gets called for type identifier 11: M_ME_NB_1.
	Scaled(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual)
	// ScaledAtMinute gets called for type identifier 12: M_ME_TB_1.
	ScaledAtMinute(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, info.CP24Time2a)
	// ScaledAtMoment gets called for type identifier 35: M_ME_TE_1.
	ScaledAtMoment(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, info.CP56Time2a)
}

type scaledProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	listener   func(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, time.Time)
	timeZone   *time.Location
	timeLeeway time.Duration
}

// ScaledProxy abstracts the variants from the ScaledMonitor interface into one
// function. Time tags are interpretated within the time-zone argument. The
// time.Time argument is always zero for type M_ME_NB_1. Time is also zero for
// Invalid() info.CP24Time2a and info.CP56Time2a.
//
// Time tags are assumed to be recent. See the example of WithinHourBefore from
// info.CP24Time2a for an explaination of the leeway setting.
// https://pkg.go.dev/github.com/pascaldekloe/part5/info#example-CP24Time2a.WithinHourBefore
func ScaledProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](listener func(info.DataUnit[Orig, Com, Obj], Obj, int16, info.Qual, time.Time), zone *time.Location, leeway time.Duration) ScaledMonitor[Orig, Com, Obj] {
	return scaledProxy[Orig, Com, Obj]{listener, zone, leeway}
}

func (proxy scaledProxy[Orig, Com, Obj]) Scaled(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual) {
	proxy.listener(u, addr, v, q, time.Time{})
}

func (proxy scaledProxy[Orig, Com, Obj]) ScaledAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual, tag info.CP24Time2a) {
	proxy.listener(u, addr, v, q, tag.WithinHourBefore(time.Now().In(proxy.timeZone).Add(proxy.timeLeeway)))
}

func (proxy scaledProxy[Orig, Com, Obj]) ScaledAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, v int16, q info.Qual, tag info.CP56Time2a) {
	proxy.listener(u, addr, v, q, tag.Within20thCentury(proxy.timeZone))
}

// FloatMonitor consumes floating points.
type FloatMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// Float gets called for type identifier 13: M_ME_NC_1.
	Float(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual)
	// FloatAtMinute gets called for type identifier 14: M_ME_TC_1.
	FloatAtMinute(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, info.CP24Time2a)
	// FloatAtMoment gets called for type identifier 36: M_ME_TF_1.
	FloatAtMoment(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, info.CP56Time2a)
}

type floatProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	listener   func(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, time.Time)
	timeZone   *time.Location
	timeLeeway time.Duration
}

// FloatProxy abstracts the variants from the FloatMonitor interface
// into one function. Time tags are interpretated within the time-zone argument.
// The time.Time argument is always zero for type M_ME_NC_1. Time is also zero
// for Invalid() info.CP24Time2a and info.CP56Time2a.
//
// Time tags are assumed to be recent. See the example of WithinHourBefore from
// info.CP24Time2a for an explaination of the leeway setting.
// https://pkg.go.dev/github.com/pascaldekloe/part5/info#example-CP24Time2a.WithinHourBefore
func FloatProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](listener func(info.DataUnit[Orig, Com, Obj], Obj, float32, info.Qual, time.Time), zone *time.Location, leeway time.Duration) FloatMonitor[Orig, Com, Obj] {
	return floatProxy[Orig, Com, Obj]{listener, zone, leeway}
}

func (proxy floatProxy[Orig, Com, Obj]) Float(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual) {
	proxy.listener(u, addr, f, q, time.Time{})
}

func (proxy floatProxy[Orig, Com, Obj]) FloatAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual, tag info.CP24Time2a) {
	proxy.listener(u, addr, f, q, tag.WithinHourBefore(time.Now().In(proxy.timeZone).Add(proxy.timeLeeway)))
}

func (proxy floatProxy[Orig, Com, Obj]) FloatAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, f float32, q info.Qual, tag info.CP56Time2a) {
	proxy.listener(u, addr, f, q, tag.Within20thCentury(proxy.timeZone))
}

// TotalsMonitor consumes integrated totals.
type TotalsMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// Totals gets called for type identifier 15: M_IT_NA_1.
	Totals(info.DataUnit[Orig, Com, Obj], Obj, info.Counter)
	// TotalsAtMinute gets called for type identifier 16: M_IT_TA_1.
	TotalsAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.Counter, info.CP24Time2a)
	// TotalsAtMoment gets called for type identifier 37: M_IT_TB_1.
	TotalsAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.Counter, info.CP56Time2a)
}

type totalsProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	listener   func(info.DataUnit[Orig, Com, Obj], Obj, info.Counter, time.Time)
	timeZone   *time.Location
	timeLeeway time.Duration
}

// TotalsProxy abstracts the variants from the TotalsMonitor interface
// into one function. Time tags are interpretated within the time-zone argument.
// The time.Time argument is always zero for type M_IT_NA_1. Time is also zero
// for Invalid() info.CP24Time2a and info.CP56Time2a.
//
// Time tags are assumed to be recent. See the example of WithinHourBefore from
// info.CP24Time2a for an explaination of the leeway setting.
// https://pkg.go.dev/github.com/pascaldekloe/part5/info#example-CP24Time2a.WithinHourBefore
func TotalsProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](listener func(info.DataUnit[Orig, Com, Obj], Obj, info.Counter, time.Time), zone *time.Location, leeway time.Duration) TotalsMonitor[Orig, Com, Obj] {
	return totalsProxy[Orig, Com, Obj]{listener, zone, leeway}
}

func (proxy totalsProxy[Orig, Com, Obj]) Totals(u info.DataUnit[Orig, Com, Obj], addr Obj, c info.Counter) {
	proxy.listener(u, addr, c, time.Time{})
}

func (proxy totalsProxy[Orig, Com, Obj]) TotalsAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, c info.Counter, tag info.CP24Time2a) {
	proxy.listener(u, addr, c, tag.WithinHourBefore(time.Now().In(proxy.timeZone).Add(proxy.timeLeeway)))
}

func (proxy totalsProxy[Orig, Com, Obj]) TotalsAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, c info.Counter, tag info.CP56Time2a) {
	proxy.listener(u, addr, c, tag.Within20thCentury(proxy.timeZone))
}

// ProtectMonitor consumes events from protection equipment.
// The uint16 with the ellapsed time should contain milliseconds in
// range 0..60000. See info.EllapsedTimeInvalid in the DoublePtQual
// before using such value.
type ProtectMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// ProtectAtMinute gets called for type identifier 17: M_EP_TA_1.
	ProtectAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, uint16, info.CP24Time2a)
	// ProtectAtMoment gets called for type identifier 38: M_EP_TD_1.
	ProtectAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, uint16, info.CP56Time2a)
}

type protEquipProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	listener   func(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, uint16, time.Time)
	timeZone   *time.Location
	timeLeeway time.Duration
}

// ProtectProxy abstracts the variants from the ProtectMonitor interface into
// one function. Time tags are interpretated within the time-zone argument. Time
// is zero for Invalid() info.CP24Time2a and info.CP56Time2a.
//
// Time tags are assumed to be recent. See the example of WithinHourBefore from
// info.CP24Time2a for an explaination of the leeway setting.
// https://pkg.go.dev/github.com/pascaldekloe/part5/info#example-CP24Time2a.WithinHourBefore
func ProtectProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](
	listener func(info.DataUnit[Orig, Com, Obj], Obj, info.DoublePtQual, uint16, time.Time),
	zone *time.Location, leeway time.Duration) ProtectMonitor[Orig, Com, Obj] {
	return protEquipProxy[Orig, Com, Obj]{listener, zone, leeway}
}

func (proxy protEquipProxy[Orig, Com, Obj]) ProtectAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, pt info.DoublePtQual, ms uint16, tag info.CP24Time2a) {
	proxy.listener(u, addr, pt, ms, tag.WithinHourBefore(time.Now().In(proxy.timeZone).Add(proxy.timeLeeway)))
}

func (proxy protEquipProxy[Orig, Com, Obj]) ProtectAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, pt info.DoublePtQual, ms uint16, tag info.CP56Time2a) {
	proxy.listener(u, addr, pt, ms, tag.Within20thCentury(proxy.timeZone))
}

// ProtectStartMonitor consumes start events from protection equipment.
// The uint16 with the relay duration time should contain milliseconds in
// range 0..60000. See info.EllapsedTimeInvalid before using such value.
// Quality descriptor info.Overflow does not apply.
type ProtectStartMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// ProtectStartAtMinute gets called for type identifier 18: M_EP_TB_1
	ProtectStartAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.ProtectStart, info.Qual, uint16, info.CP24Time2a)
	// ProtectStartAtMoment for type identifier 39: M_EP_TE_1
	ProtectStartAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.ProtectStart, info.Qual, uint16, info.CP56Time2a)
}

type protEquipStartProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	listener   func(info.DataUnit[Orig, Com, Obj], Obj, info.ProtectStart, info.Qual, uint16, time.Time)
	timeZone   *time.Location
	timeLeeway time.Duration
}

// ProtectStartProxy abstracts the variants from the ProtectStartMonitor
// interface into one function. Time tags are interpretated within the time-zone
// argument. Time is zero for Invalid() info.CP24Time2a and info.CP56Time2a.
//
// Time tags are assumed to be recent. See the example of WithinHourBefore from
// info.CP24Time2a for an explaination of the leeway setting.
// https://pkg.go.dev/github.com/pascaldekloe/part5/info#example-CP24Time2a.WithinHourBefore
func ProtectStartProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](
	listener func(info.DataUnit[Orig, Com, Obj], Obj, info.ProtectStart, info.Qual, uint16, time.Time),
	zone *time.Location, leeway time.Duration) ProtectStartMonitor[Orig, Com, Obj] {
	return protEquipStartProxy[Orig, Com, Obj]{listener, zone, leeway}
}

func (proxy protEquipStartProxy[Orig, Com, Obj]) ProtectStartAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, start info.ProtectStart, q info.Qual, ms uint16, tag info.CP24Time2a) {
	proxy.listener(u, addr, start, q, ms, tag.WithinHourBefore(time.Now().In(proxy.timeZone).Add(proxy.timeLeeway)))
}

func (proxy protEquipStartProxy[Orig, Com, Obj]) ProtectStartAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, start info.ProtectStart, q info.Qual, ms uint16, tag info.CP56Time2a) {
	proxy.listener(u, addr, start, q, ms, tag.Within20thCentury(proxy.timeZone))
}

// ProtectOutMonitor consumes command output form protection equipment.
// The uint16 with the relay operating time should contain milliseconds in
// range 0..60000. See info.EllapsedTimeInvalid before using such value.
// Quality descriptor info.Overflow does not apply.
type ProtectOutMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// ProtectOutAtMinute for type identifier 19: M_EP_TC_1
	ProtectOutAtMinute(info.DataUnit[Orig, Com, Obj], Obj, info.ProtectOut, info.Qual, uint16, info.CP24Time2a)
	// ProtectOutAtMoment for type identifier 40: M_EP_TF_1
	ProtectOutAtMoment(info.DataUnit[Orig, Com, Obj], Obj, info.ProtectOut, info.Qual, uint16, info.CP56Time2a)
}

type protEquipOutProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] struct {
	listener   func(info.DataUnit[Orig, Com, Obj], Obj, info.ProtectOut, info.Qual, uint16, time.Time)
	timeZone   *time.Location
	timeLeeway time.Duration
}

// ProtectOutProxy abstracts the variants from the ProtectOutMonitor interface
// into one function. Time tags are interpretated within the time-zone argument.
// Time is zero for Invalid() info.CP24Time2a and info.CP56Time2a.
//
// Time tags are assumed to be recent. See the example of WithinHourBefore from
// info.CP24Time2a for an explaination of the leeway setting.
// https://pkg.go.dev/github.com/pascaldekloe/part5/info#example-CP24Time2a.WithinHourBefore
func ProtectOutProxy[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](
	listener func(info.DataUnit[Orig, Com, Obj], Obj, info.ProtectOut, info.Qual, uint16, time.Time),
	zone *time.Location, leeway time.Duration) ProtectOutMonitor[Orig, Com, Obj] {
	return protEquipOutProxy[Orig, Com, Obj]{listener, zone, leeway}
}

func (proxy protEquipOutProxy[Orig, Com, Obj]) ProtectOutAtMinute(u info.DataUnit[Orig, Com, Obj], addr Obj, out info.ProtectOut, q info.Qual, ms uint16, tag info.CP24Time2a) {
	proxy.listener(u, addr, out, q, ms, tag.WithinHourBefore(time.Now().In(proxy.timeZone).Add(proxy.timeLeeway)))
}

func (proxy protEquipOutProxy[Orig, Com, Obj]) ProtectOutAtMoment(u info.DataUnit[Orig, Com, Obj], addr Obj, out info.ProtectOut, q info.Qual, ms uint16, tag info.CP56Time2a) {
	proxy.listener(u, addr, out, q, ms, tag.Within20thCentury(proxy.timeZone))
}

// InitEndMonitor consumes end-of-initialization notification.
type InitEndMonitor[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr] interface {
	// Totals gets called for type identifier 70: M_EI_NA_1.
	InitEnd(info.DataUnit[Orig, Com, Obj], info.InitCause)
}

// MonitorDataUnit has two errors for the selection on type identifier.
var (
	// ErrNotMonitor rejects an info.DataUnit based on its type identifier.
	ErrNotMonitor = errors.New("part5: ASDU type identifier not in monitor information range 1..44")

	// ErrMonitorReserve signals type extension within monitor information range 1..44.
	ErrMonitorReserve = errors.New("part5: ASDU type identifier reserved for further compatible definitions")
)

var errInfoSize = errors.New("part5: size of ASDU payload doesn't match the variable structure qualifier")

// MonitorDataUnit propagates information objects in u to the corresponding
// listener method from mon, filtering with ErrNotMontior and ErrMonitorReserve.
// DataUnits with no [zero] information elements pass without invocation to mon.
func MonitorDataUnit[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](mon Monitor[Orig, Com, Obj], u info.DataUnit[Orig, Com, Obj]) error {
	// monitor type identifiers (M_*) are in range 1..44
	if u.Type-1 > 43 {
		return ErrNotMonitor
	}

	// NOTE: Go can't get the array length from a generic as a constant yet.
	var addr Obj

	switch u.Type {
	case info.M_SP_NA_1: // single-point
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 1)
			if err != nil {
				return err
			}
			for _, b := range u.Info[len(addr):] {
				mon.SinglePt(u, addr, info.SinglePtQual(b))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+1) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+1 <= len(u.Info); i += len(addr) + 1 {
				mon.SinglePt(u,
					Obj(u.Info[i:i+len(addr)]),
					info.SinglePtQual(u.Info[i+len(addr)]),
				)
			}
		}

	case info.M_SP_TA_1: // single-point with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_SP_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+4 <= len(u.Info); i += len(addr) + 4 {
			mon.SinglePtAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				info.SinglePtQual(u.Info[i+len(addr)]),
				info.CP24Time2a(u.Info[i+len(addr)+1:i+len(addr)+4]),
			)
		}

	case info.M_SP_TB_1: // single-point with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_SP_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+8 <= len(u.Info); i += len(addr) + 8 {
			mon.SinglePtAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				info.SinglePtQual(u.Info[i+len(addr)]),
				info.CP56Time2a(u.Info[i+len(addr)+1:i+len(addr)+8]),
			)
		}

	case info.M_PS_NA_1: // single-points with status change
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 5)
			if err != nil {
				return err
			}
			for i := len(addr); i+5 <= len(u.Info); i += 5 {
				mon.SinglePtChangePack(u, addr,
					info.SinglePtChangePack(
						binary.BigEndian.Uint32(
							u.Info[i:i+4],
						),
					),
					info.Qual(u.Info[i+4]),
				)
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+5 <= len(u.Info); i += len(addr) + 5 {
				mon.SinglePtChangePack(u,
					Obj(u.Info[i:i+len(addr)]),
					info.SinglePtChangePack(
						binary.BigEndian.Uint32(
							u.Info[i+len(addr):i+len(addr)+4],
						),
					),
					info.Qual(u.Info[i+len(addr)+4]),
				)
			}
		}

	case info.M_DP_NA_1: // double-point
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 1)
			if err != nil {
				return err
			}
			for _, b := range u.Info[len(addr):] {
				mon.DoublePt(u, addr, info.DoublePtQual(b))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+1) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+1 <= len(u.Info); i += len(addr) + 1 {
				mon.DoublePt(u,
					Obj(u.Info[i:i+len(addr)]),
					info.DoublePtQual(u.Info[i+len(addr)]),
				)
			}
		}

	case info.M_DP_TA_1: // double-point with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_DP_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+4) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+4 <= len(u.Info); i += len(addr) + 4 {
			mon.DoublePtAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				info.DoublePtQual(u.Info[i+len(addr)]),
				info.CP24Time2a(u.Info[i+len(addr)+1:i+len(addr)+4]),
			)
		}

	case info.M_DP_TB_1: // double-point with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_DP_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+8 <= len(u.Info); i += len(addr) + 8 {
			mon.DoublePtAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				info.DoublePtQual(u.Info[i+len(addr)]),
				info.CP56Time2a(u.Info[i+len(addr)+1:i+len(addr)+8]),
			)
		}

	case info.M_ST_NA_1: // step position
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 2)
			if err != nil {
				return err
			}
			for i := len(addr); i+2 <= len(u.Info); i += 2 {
				mon.Step(u, addr, info.StepQual(u.Info[i:i+2]))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+2) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+2 <= len(u.Info); i += len(addr) + 2 {
				mon.Step(u,
					Obj(u.Info[i:i+len(addr)]),
					info.StepQual(u.Info[i+len(addr):i+len(addr)+2]),
				)
			}
		}

	case info.M_ST_TA_1: // step position with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ST_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+5) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+5 <= len(u.Info); i += len(addr) + 5 {
			mon.StepAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				info.StepQual(u.Info[i+len(addr):i+len(addr)+2]),
				info.CP24Time2a(u.Info[i+len(addr)+2:i+len(addr)+5]),
			)
		}

	case info.M_ST_TB_1: // step position with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ST_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+9) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+9 <= len(u.Info); i += len(addr) + 9 {
			mon.StepAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				info.StepQual(u.Info[i+len(addr):i+len(addr)+2]),
				info.CP56Time2a(u.Info[i+len(addr)+2:i+len(addr)+9]),
			)
		}

	case info.M_BO_NA_1: // bit string
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 5)
			if err != nil {
				return err
			}
			for i := len(addr); i+5 <= len(u.Info); i += 5 {
				mon.Bits(u, addr, info.BitsQual(u.Info[i:i+5]))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+5 <= len(u.Info); i += len(addr) + 5 {
				mon.Bits(u,
					Obj(u.Info[i:i+len(addr)]),
					info.BitsQual(u.Info[i+len(addr):i+len(addr)+5]),
				)
			}
		}

	case info.M_BO_TA_1: // bit string with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_BO_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+8 <= len(u.Info); i += len(addr) + 8 {
			mon.BitsAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				info.BitsQual(u.Info[i+len(addr):i+len(addr)+5]),
				info.CP24Time2a(u.Info[i+len(addr)+5:i+len(addr)+8]),
			)
		}

	case info.M_BO_TB_1: // bit string with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_BO_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+12 <= len(u.Info); i += len(addr) + 12 {
			mon.BitsAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				info.BitsQual(u.Info[i+len(addr):i+len(addr)+5]),
				info.CP56Time2a(u.Info[i+len(addr)+5:i+len(addr)+12]),
			)
		}

	case info.M_ME_ND_1: // normalized value without quality descriptor
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 2)
			if err != nil {
				return err
			}
			for i := len(addr); i+1 < len(u.Info); i += 2 {
				mon.NormUnqual(u, addr, info.Norm(u.Info[i:i+2]))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+2) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+2 <= len(u.Info); i += len(addr) + 2 {
				mon.NormUnqual(u,
					Obj(u.Info[i:i+len(addr)]),
					info.Norm(u.Info[i+len(addr):i+len(addr)+2]),
				)
			}
		}

	case info.M_ME_NA_1: // normalized value
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 3)
			if err != nil {
				return err
			}
			for i := len(addr); i+3 <= len(u.Info); i += 3 {
				mon.Norm(u, addr, info.NormQual(u.Info[i:i+3]))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+3) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+3 <= len(u.Info); i += len(addr) + 3 {
				mon.Norm(u,
					Obj(u.Info[i:i+len(addr)]),
					info.NormQual(u.Info[i+len(addr):i+len(addr)+3]),
				)
			}
		}

	case info.M_ME_TA_1: // normalized value with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+6 <= len(u.Info); i += len(addr) + 6 {
			mon.NormAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				info.NormQual(u.Info[i+len(addr):i+len(addr)+3]),
				info.CP24Time2a(u.Info[i+len(addr)+3:i+len(addr)+6]),
			)
		}

	case info.M_ME_TD_1: // normalized value with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TD_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+10 <= len(u.Info); i += len(addr) + 10 {
			mon.NormAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				info.NormQual(u.Info[i+len(addr):i+len(addr)+3]),
				info.CP56Time2a(u.Info[i+len(addr)+3:i+len(addr)+10]),
			)
		}

	case info.M_ME_NB_1: // scaled value
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 3)
			if err != nil {
				return err
			}
			for i := len(addr); i+3 <= len(u.Info); i += 3 {
				mon.Scaled(u, addr,
					int16(
						binary.LittleEndian.Uint16(
							u.Info[i:i+2],
						),
					),
					info.Qual(u.Info[i+2]),
				)
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+3) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+3 <= len(u.Info); i += len(addr) + 3 {
				mon.Scaled(u,
					Obj(u.Info[i:i+len(addr)]),
					int16(
						binary.LittleEndian.Uint16(
							u.Info[i+len(addr):i+len(addr)+2],
						),
					),
					info.Qual(u.Info[i+len(addr)+2]),
				)
			}
		}

	case info.M_ME_TB_1: // scaled value with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+6) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+6 <= len(u.Info); i += len(addr) + 6 {
			mon.ScaledAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				int16(
					binary.LittleEndian.Uint16(
						u.Info[i+len(addr):i+len(addr)+2],
					),
				),
				info.Qual(u.Info[i+len(addr)+2]),
				info.CP24Time2a(u.Info[i+len(addr)+3:i+len(addr)+6]),
			)
		}

	case info.M_ME_TE_1: // scaled value with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TE_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+10) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+10 <= len(u.Info); i += len(addr) + 10 {
			mon.ScaledAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				int16(
					binary.LittleEndian.Uint16(
						u.Info[i+len(addr):i+len(addr)+2],
					),
				),
				info.Qual(u.Info[i+len(addr)+2]),
				info.CP56Time2a(u.Info[i+len(addr)+3:i+len(addr)+10]),
			)
			u.Info = u.Info[len(addr)+10:]
		}

	case info.M_ME_NC_1: // floating-point
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 5)
			if err != nil {
				return err
			}
			for i := len(addr); i+5 <= len(u.Info); i += 5 {
				mon.Float(u, addr,
					math.Float32frombits(
						binary.LittleEndian.Uint32(
							u.Info[i:i+4],
						),
					),
					info.Qual(u.Info[i+4]),
				)
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+5 <= len(u.Info); i += len(addr) + 5 {
				mon.Float(u,
					Obj(u.Info[i:i+len(addr)]),
					math.Float32frombits(
						binary.LittleEndian.Uint32(
							u.Info[i+len(addr):i+len(addr)+4],
						),
					),
					info.Qual(u.Info[i+len(addr)+4]),
				)
			}
		}

	case info.M_ME_TC_1: // floating point with 3 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TC_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+8 <= len(u.Info); i += len(addr) + 8 {
			mon.FloatAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						u.Info[i+len(addr):i+len(addr)+4],
					),
				),
				info.Qual(u.Info[i+len(addr)+4]),
				info.CP24Time2a(u.Info[i+len(addr)+5:i+len(addr)+8]),
			)
		}

	case info.M_ME_TF_1: // floating point with 7 octet time-tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_ME_TF_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+12 <= len(u.Info); i += len(addr) + 12 {
			mon.FloatAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				math.Float32frombits(
					binary.LittleEndian.Uint32(
						u.Info[i+len(addr):i+len(addr)+4],
					),
				),
				info.Qual(u.Info[i+len(addr)+4]),
				info.CP56Time2a(u.Info[i+len(addr)+5:i+len(addr)+12]),
			)
		}

	case info.M_IT_NA_1: // integrated totals.
		if u.Enc.AddrSeq() {
			addr, err := addrSeqStart(&u, 5)
			if err != nil {
				return err
			}
			for i := len(addr); i+5 <= len(u.Info); i += 5 {
				mon.Totals(u, addr, info.Counter(u.Info[i:i+5]))
				addr, _ = u.Params.ObjAddrN(addr.N() + 1)
			}
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+5) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+5 <= len(u.Info); i += len(addr) + 5 {
				mon.Totals(u, Obj(u.Info[i:i+len(addr)]),
					info.Counter(u.Info[i+len(addr):i+len(addr)+5]),
				)
			}
		}

	case info.M_IT_TA_1: // integrated totals with 3 octet time-tag.
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_IT_TA_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+8) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+8 <= len(u.Info); i += len(addr) + 8 {
			mon.TotalsAtMinute(u,
				Obj(u.Info[i:i+len(addr)]),
				info.Counter(u.Info[i+len(addr):i+len(addr)+5]),
				info.CP24Time2a(u.Info[i+len(addr)+5:i+len(addr)+8]),
			)
		}

	case info.M_IT_TB_1: // integrated totals with 7 octet time-tag.
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_IT_TB_1 not allowed")
		}
		if len(u.Info) != u.Enc.Count()*(len(addr)+12) {
			return errInfoSize
		}
		for i := 0; i+len(addr)+12 <= len(u.Info); i += len(addr) + 12 {
			mon.TotalsAtMoment(u,
				Obj(u.Info[i:i+len(addr)]),
				info.Counter(u.Info[i+len(addr):i+len(addr)+5]),
				info.CP56Time2a(u.Info[i+len(addr)+5:i+len(addr)+12]),
			)
		}

	case info.M_EP_TA_1: // protection equipment with 3-octet time tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_EP_TA_1 not allowed")
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+6) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+6 <= len(u.Info); i += len(addr) + 6 {
				mon.ProtectAtMinute(u,
					Obj(u.Info[i:i+len(addr)]),
					info.DoublePtQual(u.Info[i+len(addr)]),
					binary.BigEndian.Uint16(u.Info[i+len(addr)+1:i+len(addr)+3]),
					info.CP24Time2a(u.Info[i+len(addr)+3:i+len(addr)+6]),
				)
			}
		}

	case info.M_EP_TD_1: // protection equipment with 7-octet time tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_EP_TD_1 not allowed")
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+10) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+10 <= len(u.Info); i += len(addr) + 10 {
				mon.ProtectAtMoment(u,
					Obj(u.Info[i:i+len(addr)]),
					info.DoublePtQual(u.Info[i+len(addr)]),
					binary.BigEndian.Uint16(u.Info[i+len(addr)+1:i+len(addr)+3]),
					info.CP56Time2a(u.Info[i+len(addr)+3:i+len(addr)+10]),
				)
			}
		}

	case info.M_EP_TB_1: // start of protection equipment with 3-octet time tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_EP_TB_1 not allowed")
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+7) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+7 <= len(u.Info); i += len(addr) + 7 {
				mon.ProtectStartAtMinute(u,
					Obj(u.Info[i:i+len(addr)]),
					info.ProtectStart(u.Info[i+len(addr)]),
					info.Qual(u.Info[i+len(addr)+1]),
					binary.BigEndian.Uint16(u.Info[i+len(addr)+2:i+len(addr)+4]),
					info.CP24Time2a(u.Info[i+len(addr)+4:i+len(addr)+7]),
				)
			}
		}

	case info.M_EP_TE_1: // start of protection equipment with 7-octet time tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_EP_TE_1 not allowed")
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+11) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+11 <= len(u.Info); i += len(addr) + 11 {
				mon.ProtectStartAtMoment(u,
					Obj(u.Info[i:i+len(addr)]),
					info.ProtectStart(u.Info[i+len(addr)]),
					info.Qual(u.Info[i+len(addr)+1]),
					binary.BigEndian.Uint16(u.Info[i+len(addr)+2:i+len(addr)+4]),
					info.CP56Time2a(u.Info[i+len(addr)+4:i+len(addr)+11]),
				)
			}
		}

	case info.M_EP_TC_1: // output of protection equipment with 3-octet time tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_EP_TC_1 not allowed")
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+7) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+7 <= len(u.Info); i += len(addr) + 7 {
				mon.ProtectOutAtMinute(u,
					Obj(u.Info[i:i+len(addr)]),
					info.ProtectOut(u.Info[i+len(addr)]),
					info.Qual(u.Info[i+len(addr)+1]),
					binary.BigEndian.Uint16(u.Info[i+len(addr)+2:i+len(addr)+4]),
					info.CP24Time2a(u.Info[i+len(addr)+4:i+len(addr)+7]),
				)
			}
		}

	case info.M_EP_TF_1: // output of protection equipment with 7-octet time tag
		if u.Enc.AddrSeq() {
			return errors.New("part5: ASDU address sequence with M_EP_TF_1 not allowed")
		} else {
			if len(u.Info) != u.Enc.Count()*(len(addr)+11) {
				return errInfoSize
			}
			for i := 0; i+len(addr)+11 <= len(u.Info); i += len(addr) + 11 {
				mon.ProtectOutAtMoment(u,
					Obj(u.Info[i:i+len(addr)]),
					info.ProtectOut(u.Info[i+len(addr)]),
					info.Qual(u.Info[i+len(addr)+1]),
					binary.BigEndian.Uint16(u.Info[i+len(addr)+2:i+len(addr)+4]),
					info.CP56Time2a(u.Info[i+len(addr)+4:i+len(addr)+11]),
				)
			}
		}

	case info.M_EI_NA_1: // end of initialization
		if len(u.Info) != len(addr)+1 {
			return errInfoSize
		}
		mon.InitEnd(u, info.InitCause(u.Info[len(addr)]))

	default:
		return ErrMonitorReserve
	}
	return nil
}

func addrSeqStart[Orig info.OrigAddr, Com info.ComAddr, Obj info.ObjAddr](u *info.DataUnit[Orig, Com, Obj], encSize int) (addr Obj, err error) {
	if len(u.Info) != len(addr)+u.Enc.Count()*encSize {
		return addr, errInfoSize
	}
	addr = Obj(u.Info[:len(addr)])

	// overflow check
	lastN := addr.N() + uint(u.Enc.Count()) - 1
	if _, ok := u.Params.ObjAddrN(lastN); !ok {
		return addr, info.ErrAddrSeq
	}
	return addr, nil
}
