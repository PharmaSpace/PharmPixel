// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package service

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonB499ec80DecodePixelStoreService(in *jlexer.Lexer, out *UniFarmReceiptResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "ОтветНаДанныеПоЗапросу":
			(out.Response).UnmarshalEasyJSON(in)
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonB499ec80EncodePixelStoreService(out *jwriter.Writer, in UniFarmReceiptResponse) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"ОтветНаДанныеПоЗапросу\":"
		out.RawString(prefix[1:])
		(in.Response).MarshalEasyJSON(out)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UniFarmReceiptResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonB499ec80EncodePixelStoreService(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UniFarmReceiptResponse) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonB499ec80EncodePixelStoreService(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UniFarmReceiptResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonB499ec80DecodePixelStoreService(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UniFarmReceiptResponse) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonB499ec80DecodePixelStoreService(l, v)
}
func easyjsonB499ec80DecodePixelStoreService1(in *jlexer.Lexer, out *UniFarmReceipt) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "Период":
			out.Date = string(in.String())
		case "ТоварКод":
			out.ProductCode = string(in.String())
		case "ТоварНаименование":
			out.ProductName = string(in.String())
		case "Партия":
			out.PartNumber = string(in.String())
		case "ЦенаЗакуп":
			out.PriceSellIn = string(in.String())
		case "ЦенаРозн":
			out.PriceSellOut = string(in.String())
		case "Серия":
			out.Serial = string(in.String())
		case "ПроизводительНаименование":
			out.ManufacturerName = string(in.String())
		case "ПоставщикНаименование":
			out.ProviderName = string(in.String())
		case "ПоставщикИНН":
			out.InnProvider = string(in.String())
		case "ЗаводскойНомерККТ":
			out.NumberKKT = string(in.String())
		case "СкладНаименование":
			out.WarehouseName = string(in.String())
		case "СуммаСкидки":
			out.DiscountPrice = string(in.String())
		case "Количество":
			out.Quantity = string(in.String())
		case "СуммаЗакуп":
			out.SumPriceSellIn = string(in.String())
		case "СуммаРозн":
			out.SumPriceSellOut = string(in.String())
		case "НомерЧека":
			out.DocumentNumber = string(in.String())
		case "НомерФискальногоДокумента":
			out.NumberFiscalDocument = string(in.String())
		case "ЗаводскойНомерФН":
			out.ManufNumberFNKKT = string(in.String())
		case "ФискальныеДанные":
			out.FiscalData = string(in.String())
		case "СуммаЧека":
			out.SumReceipt = string(in.String())
		case "Штрихкод":
			out.Barcode = string(in.String())
		case "СрокГодности":
			out.ExpireDate = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonB499ec80EncodePixelStoreService1(out *jwriter.Writer, in UniFarmReceipt) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"Период\":"
		out.RawString(prefix[1:])
		out.String(string(in.Date))
	}
	{
		const prefix string = ",\"ТоварКод\":"
		out.RawString(prefix)
		out.String(string(in.ProductCode))
	}
	{
		const prefix string = ",\"ТоварНаименование\":"
		out.RawString(prefix)
		out.String(string(in.ProductName))
	}
	{
		const prefix string = ",\"Партия\":"
		out.RawString(prefix)
		out.String(string(in.PartNumber))
	}
	{
		const prefix string = ",\"ЦенаЗакуп\":"
		out.RawString(prefix)
		out.String(string(in.PriceSellIn))
	}
	{
		const prefix string = ",\"ЦенаРозн\":"
		out.RawString(prefix)
		out.String(string(in.PriceSellOut))
	}
	{
		const prefix string = ",\"Серия\":"
		out.RawString(prefix)
		out.String(string(in.Serial))
	}
	{
		const prefix string = ",\"ПроизводительНаименование\":"
		out.RawString(prefix)
		out.String(string(in.ManufacturerName))
	}
	{
		const prefix string = ",\"ПоставщикНаименование\":"
		out.RawString(prefix)
		out.String(string(in.ProviderName))
	}
	{
		const prefix string = ",\"ПоставщикИНН\":"
		out.RawString(prefix)
		out.String(string(in.InnProvider))
	}
	{
		const prefix string = ",\"ЗаводскойНомерККТ\":"
		out.RawString(prefix)
		out.String(string(in.NumberKKT))
	}
	{
		const prefix string = ",\"СкладНаименование\":"
		out.RawString(prefix)
		out.String(string(in.WarehouseName))
	}
	{
		const prefix string = ",\"СуммаСкидки\":"
		out.RawString(prefix)
		out.String(string(in.DiscountPrice))
	}
	{
		const prefix string = ",\"Количество\":"
		out.RawString(prefix)
		out.String(string(in.Quantity))
	}
	{
		const prefix string = ",\"СуммаЗакуп\":"
		out.RawString(prefix)
		out.String(string(in.SumPriceSellIn))
	}
	{
		const prefix string = ",\"СуммаРозн\":"
		out.RawString(prefix)
		out.String(string(in.SumPriceSellOut))
	}
	{
		const prefix string = ",\"НомерЧека\":"
		out.RawString(prefix)
		out.String(string(in.DocumentNumber))
	}
	{
		const prefix string = ",\"НомерФискальногоДокумента\":"
		out.RawString(prefix)
		out.String(string(in.NumberFiscalDocument))
	}
	{
		const prefix string = ",\"ЗаводскойНомерФН\":"
		out.RawString(prefix)
		out.String(string(in.ManufNumberFNKKT))
	}
	{
		const prefix string = ",\"ФискальныеДанные\":"
		out.RawString(prefix)
		out.String(string(in.FiscalData))
	}
	{
		const prefix string = ",\"СуммаЧека\":"
		out.RawString(prefix)
		out.String(string(in.SumReceipt))
	}
	{
		const prefix string = ",\"Штрихкод\":"
		out.RawString(prefix)
		out.String(string(in.Barcode))
	}
	{
		const prefix string = ",\"СрокГодности\":"
		out.RawString(prefix)
		out.String(string(in.ExpireDate))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UniFarmReceipt) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonB499ec80EncodePixelStoreService1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UniFarmReceipt) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonB499ec80EncodePixelStoreService1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UniFarmReceipt) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonB499ec80DecodePixelStoreService1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UniFarmReceipt) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonB499ec80DecodePixelStoreService1(l, v)
}
func easyjsonB499ec80DecodePixelStoreService2(in *jlexer.Lexer, out *UniFarmProductResponse) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "ОтветНаДанныеПоЗапросу":
			(out.Response).UnmarshalEasyJSON(in)
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonB499ec80EncodePixelStoreService2(out *jwriter.Writer, in UniFarmProductResponse) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"ОтветНаДанныеПоЗапросу\":"
		out.RawString(prefix[1:])
		(in.Response).MarshalEasyJSON(out)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UniFarmProductResponse) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonB499ec80EncodePixelStoreService2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UniFarmProductResponse) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonB499ec80EncodePixelStoreService2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UniFarmProductResponse) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonB499ec80DecodePixelStoreService2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UniFarmProductResponse) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonB499ec80DecodePixelStoreService2(l, v)
}
func easyjsonB499ec80DecodePixelStoreService3(in *jlexer.Lexer, out *UniFarmProduct) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "СкладНаименование":
			out.WarehouseName = string(in.String())
		case "ТоварНаименование":
			out.ProductName = string(in.String())
		case "ПроизводительНаименование":
			out.ManufacturerName = string(in.String())
		case "ПоставщикНаименование":
			out.SupplierName = string(in.String())
		case "ПоставщикИНН":
			out.SupplierINN = string(in.String())
		case "Партия":
			out.PartNumber = string(in.String())
		case "Серия":
			out.Serial = string(in.String())
		case "ДатаЧека":
			out.Date = string(in.String())
		case "Количество":
			out.Stock = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonB499ec80EncodePixelStoreService3(out *jwriter.Writer, in UniFarmProduct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"СкладНаименование\":"
		out.RawString(prefix[1:])
		out.String(string(in.WarehouseName))
	}
	{
		const prefix string = ",\"ТоварНаименование\":"
		out.RawString(prefix)
		out.String(string(in.ProductName))
	}
	{
		const prefix string = ",\"ПроизводительНаименование\":"
		out.RawString(prefix)
		out.String(string(in.ManufacturerName))
	}
	{
		const prefix string = ",\"ПоставщикНаименование\":"
		out.RawString(prefix)
		out.String(string(in.SupplierName))
	}
	{
		const prefix string = ",\"ПоставщикИНН\":"
		out.RawString(prefix)
		out.String(string(in.SupplierINN))
	}
	{
		const prefix string = ",\"Партия\":"
		out.RawString(prefix)
		out.String(string(in.PartNumber))
	}
	{
		const prefix string = ",\"Серия\":"
		out.RawString(prefix)
		out.String(string(in.Serial))
	}
	{
		const prefix string = ",\"ДатаЧека\":"
		out.RawString(prefix)
		out.String(string(in.Date))
	}
	{
		const prefix string = ",\"Количество\":"
		out.RawString(prefix)
		out.String(string(in.Stock))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UniFarmProduct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonB499ec80EncodePixelStoreService3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UniFarmProduct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonB499ec80EncodePixelStoreService3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UniFarmProduct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonB499ec80DecodePixelStoreService3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UniFarmProduct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonB499ec80DecodePixelStoreService3(l, v)
}
func easyjsonB499ec80DecodePixelStoreService4(in *jlexer.Lexer, out *UniFarm) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "Username":
			out.Username = string(in.String())
		case "Password":
			out.Password = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonB499ec80EncodePixelStoreService4(out *jwriter.Writer, in UniFarm) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"Username\":"
		out.RawString(prefix[1:])
		out.String(string(in.Username))
	}
	{
		const prefix string = ",\"Password\":"
		out.RawString(prefix)
		out.String(string(in.Password))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UniFarm) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonB499ec80EncodePixelStoreService4(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UniFarm) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonB499ec80EncodePixelStoreService4(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UniFarm) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonB499ec80DecodePixelStoreService4(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UniFarm) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonB499ec80DecodePixelStoreService4(l, v)
}
func easyjsonB499ec80DecodePixelStoreService5(in *jlexer.Lexer, out *ResponseForQueryReceipt) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "МассивДанных":
			if in.IsNull() {
				in.Skip()
				out.Products = nil
			} else {
				in.Delim('[')
				if out.Products == nil {
					if !in.IsDelim(']') {
						out.Products = make([]UniFarmReceipt, 0, 0)
					} else {
						out.Products = []UniFarmReceipt{}
					}
				} else {
					out.Products = (out.Products)[:0]
				}
				for !in.IsDelim(']') {
					var v1 UniFarmReceipt
					(v1).UnmarshalEasyJSON(in)
					out.Products = append(out.Products, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonB499ec80EncodePixelStoreService5(out *jwriter.Writer, in ResponseForQueryReceipt) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"МассивДанных\":"
		out.RawString(prefix[1:])
		if in.Products == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Products {
				if v2 > 0 {
					out.RawByte(',')
				}
				(v3).MarshalEasyJSON(out)
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ResponseForQueryReceipt) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonB499ec80EncodePixelStoreService5(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ResponseForQueryReceipt) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonB499ec80EncodePixelStoreService5(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ResponseForQueryReceipt) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonB499ec80DecodePixelStoreService5(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ResponseForQueryReceipt) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonB499ec80DecodePixelStoreService5(l, v)
}
func easyjsonB499ec80DecodePixelStoreService6(in *jlexer.Lexer, out *ResponseForQueryProduct) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "МассивДанных":
			if in.IsNull() {
				in.Skip()
				out.Products = nil
			} else {
				in.Delim('[')
				if out.Products == nil {
					if !in.IsDelim(']') {
						out.Products = make([]UniFarmProduct, 0, 0)
					} else {
						out.Products = []UniFarmProduct{}
					}
				} else {
					out.Products = (out.Products)[:0]
				}
				for !in.IsDelim(']') {
					var v4 UniFarmProduct
					(v4).UnmarshalEasyJSON(in)
					out.Products = append(out.Products, v4)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonB499ec80EncodePixelStoreService6(out *jwriter.Writer, in ResponseForQueryProduct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"МассивДанных\":"
		out.RawString(prefix[1:])
		if in.Products == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v5, v6 := range in.Products {
				if v5 > 0 {
					out.RawByte(',')
				}
				(v6).MarshalEasyJSON(out)
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ResponseForQueryProduct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonB499ec80EncodePixelStoreService6(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ResponseForQueryProduct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonB499ec80EncodePixelStoreService6(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ResponseForQueryProduct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonB499ec80DecodePixelStoreService6(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ResponseForQueryProduct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonB499ec80DecodePixelStoreService6(l, v)
}
