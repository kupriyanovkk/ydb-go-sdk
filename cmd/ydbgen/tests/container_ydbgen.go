// Code generated by ydbgen; DO NOT EDIT.

package tests

import (
	"strconv"

	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
)

var (
	_ = strconv.Itoa
	_ = ydb.StringValue
	_ = table.NewQueryParameters
)

func (c *Container) Scan(res *table.Result) (err error) {
	res.SeekItem("struct")
	for i0, n0 := 0, res.StructIn(); i0 < n0; i0++ {
		switch res.StructField(i0) {
		case "id":
			c.Struct.ID = res.OUTF8()
		case "ints":
			n1 := res.ListIn()
			xs0 := make([]int32, n1)
			for i1 := 0; i1 < n1; i1++ {
				res.ListItem(i1)
				var x0 int32
				x0 = res.OInt32()
				xs0[i1] = x0
			}
			c.Struct.Ints = xs0
			res.ListOut()
		}
	}
	res.StructOut()

	res.SeekItem("structs")
	n0 := res.ListIn()
	xs0 := make([]Foo, n0)
	for i0 := 0; i0 < n0; i0++ {
		res.ListItem(i0)
		var x0 Foo
		for i1, n1 := 0, res.StructIn(); i1 < n1; i1++ {
			switch res.StructField(i1) {
			case "id":
				x0.ID = res.OUTF8()
			case "ints":
				n2 := res.ListIn()
				xs1 := make([]int32, n2)
				for i2 := 0; i2 < n2; i2++ {
					res.ListItem(i2)
					var x1 int32
					x1 = res.OInt32()
					xs1[i2] = x1
				}
				x0.Ints = xs1
				res.ListOut()
			}
		}
		res.StructOut()
		xs0[i0] = x0
	}
	c.Structs = xs0
	res.ListOut()

	res.SeekItem("bytes")
	n1 := res.ListIn()
	xs1 := make([]byte, n1)
	for i0 := 0; i0 < n1; i0++ {
		res.ListItem(i0)
		var x0 byte
		x0 = ydbConvU32ToB(res.Uint32())
		xs1[i0] = x0
	}
	c.Bytes = xs1
	res.ListOut()

	res.SeekItem("strings")
	n2 := res.ListIn()
	xs2 := make([]string, n2)
	for i0 := 0; i0 < n2; i0++ {
		res.ListItem(i0)
		var x0 string
		x0 = string(res.String())
		xs2[i0] = x0
	}
	c.Strings = xs2
	res.ListOut()

	res.SeekItem("string")
	c.String = res.OString()

	return res.Err()
}

func (c *Container) QueryParameters() *table.QueryParameters {
	var v0 ydb.Value
	{
		var v1 ydb.Value
		{
			var v2 ydb.Value
			{
				vp0 := ydb.OptionalValue(ydb.UTF8Value(c.Struct.ID))
				v2 = vp0
			}
			var v3 ydb.Value
			{
				var list0 ydb.Value
				vs0 := make([]ydb.Value, len(c.Struct.Ints))
				for i0, item0 := range c.Struct.Ints {
					var v4 ydb.Value
					{
						vp0 := ydb.OptionalValue(ydb.Int32Value(item0))
						v4 = vp0
					}
					vs0[i0] = v4
				}
				if len(vs0) == 0 {
					var t1 ydb.Type
					{
						tp0 := ydb.TypeInt32
						t1 = ydb.Optional(tp0)
					}
					t0 := ydb.List(t1)
					list0 = ydb.ZeroValue(t0)
				} else {
					list0 = ydb.ListValue(vs0...)
				}
				v3 = list0
			}
			v1 = ydb.StructValue(
				ydb.StructFieldValue("id", v2),
				ydb.StructFieldValue("ints", v3),
			)
		}
		v0 = v1
	}
	var v1 ydb.Value
	{
		var list0 ydb.Value
		vs0 := make([]ydb.Value, len(c.Structs))
		for i0, item0 := range c.Structs {
			var v2 ydb.Value
			{
				var v3 ydb.Value
				{
					var v4 ydb.Value
					{
						vp0 := ydb.OptionalValue(ydb.UTF8Value(item0.ID))
						v4 = vp0
					}
					var v5 ydb.Value
					{
						var list1 ydb.Value
						vs1 := make([]ydb.Value, len(item0.Ints))
						for i1, item1 := range item0.Ints {
							var v6 ydb.Value
							{
								vp0 := ydb.OptionalValue(ydb.Int32Value(item1))
								v6 = vp0
							}
							vs1[i1] = v6
						}
						if len(vs1) == 0 {
							var t1 ydb.Type
							{
								tp0 := ydb.TypeInt32
								t1 = ydb.Optional(tp0)
							}
							t0 := ydb.List(t1)
							list1 = ydb.ZeroValue(t0)
						} else {
							list1 = ydb.ListValue(vs1...)
						}
						v5 = list1
					}
					v3 = ydb.StructValue(
						ydb.StructFieldValue("id", v4),
						ydb.StructFieldValue("ints", v5),
					)
				}
				v2 = v3
			}
			vs0[i0] = v2
		}
		if len(vs0) == 0 {
			var t1 ydb.Type
			{
				var t2 ydb.Type
				{
					fs0 := make([]ydb.StructOption, 2)
					var t3 ydb.Type
					{
						tp0 := ydb.TypeUTF8
						t3 = ydb.Optional(tp0)
					}
					fs0[0] = ydb.StructField("id", t3)
					var t4 ydb.Type
					{
						var t6 ydb.Type
						{
							tp0 := ydb.TypeInt32
							t6 = ydb.Optional(tp0)
						}
						t5 := ydb.List(t6)
						t4 = t5
					}
					fs0[1] = ydb.StructField("ints", t4)
					t2 = ydb.Struct(fs0...)
				}
				t1 = t2
			}
			t0 := ydb.List(t1)
			list0 = ydb.ZeroValue(t0)
		} else {
			list0 = ydb.ListValue(vs0...)
		}
		v1 = list0
	}
	var v2 ydb.Value
	{
		var list0 ydb.Value
		vs0 := make([]ydb.Value, len(c.Bytes))
		for i0, item0 := range c.Bytes {
			var v3 ydb.Value
			{
				vp0 := ydb.Uint32Value(uint32(item0))
				v3 = vp0
			}
			vs0[i0] = v3
		}
		if len(vs0) == 0 {
			var t1 ydb.Type
			{
				tp0 := ydb.TypeUint32
				t1 = tp0
			}
			t0 := ydb.List(t1)
			list0 = ydb.ZeroValue(t0)
		} else {
			list0 = ydb.ListValue(vs0...)
		}
		v2 = list0
	}
	var v3 ydb.Value
	{
		var list0 ydb.Value
		vs0 := make([]ydb.Value, len(c.Strings))
		for i0, item0 := range c.Strings {
			var v4 ydb.Value
			{
				vp0 := ydb.StringValue([]uint8(item0))
				v4 = vp0
			}
			vs0[i0] = v4
		}
		if len(vs0) == 0 {
			var t1 ydb.Type
			{
				tp0 := ydb.TypeString
				t1 = tp0
			}
			t0 := ydb.List(t1)
			list0 = ydb.ZeroValue(t0)
		} else {
			list0 = ydb.ListValue(vs0...)
		}
		v3 = list0
	}
	var v4 ydb.Value
	{
		vp0 := ydb.OptionalValue(ydb.StringValue(c.String))
		v4 = vp0
	}
	return table.NewQueryParameters(
		table.ValueParam("$struct", v0),
		table.ValueParam("$structs", v1),
		table.ValueParam("$bytes", v2),
		table.ValueParam("$strings", v3),
		table.ValueParam("$string", v4),
	)
}

func (c *Container) StructValue() ydb.Value {
	var v0 ydb.Value
	{
		var v1 ydb.Value
		{
			var v2 ydb.Value
			{
				var v3 ydb.Value
				{
					vp0 := ydb.OptionalValue(ydb.UTF8Value(c.Struct.ID))
					v3 = vp0
				}
				var v4 ydb.Value
				{
					var list0 ydb.Value
					vs0 := make([]ydb.Value, len(c.Struct.Ints))
					for i0, item0 := range c.Struct.Ints {
						var v5 ydb.Value
						{
							vp0 := ydb.OptionalValue(ydb.Int32Value(item0))
							v5 = vp0
						}
						vs0[i0] = v5
					}
					if len(vs0) == 0 {
						var t1 ydb.Type
						{
							tp0 := ydb.TypeInt32
							t1 = ydb.Optional(tp0)
						}
						t0 := ydb.List(t1)
						list0 = ydb.ZeroValue(t0)
					} else {
						list0 = ydb.ListValue(vs0...)
					}
					v4 = list0
				}
				v2 = ydb.StructValue(
					ydb.StructFieldValue("id", v3),
					ydb.StructFieldValue("ints", v4),
				)
			}
			v1 = v2
		}
		var v2 ydb.Value
		{
			var list0 ydb.Value
			vs0 := make([]ydb.Value, len(c.Structs))
			for i0, item0 := range c.Structs {
				var v3 ydb.Value
				{
					var v4 ydb.Value
					{
						var v5 ydb.Value
						{
							vp0 := ydb.OptionalValue(ydb.UTF8Value(item0.ID))
							v5 = vp0
						}
						var v6 ydb.Value
						{
							var list1 ydb.Value
							vs1 := make([]ydb.Value, len(item0.Ints))
							for i1, item1 := range item0.Ints {
								var v7 ydb.Value
								{
									vp0 := ydb.OptionalValue(ydb.Int32Value(item1))
									v7 = vp0
								}
								vs1[i1] = v7
							}
							if len(vs1) == 0 {
								var t1 ydb.Type
								{
									tp0 := ydb.TypeInt32
									t1 = ydb.Optional(tp0)
								}
								t0 := ydb.List(t1)
								list1 = ydb.ZeroValue(t0)
							} else {
								list1 = ydb.ListValue(vs1...)
							}
							v6 = list1
						}
						v4 = ydb.StructValue(
							ydb.StructFieldValue("id", v5),
							ydb.StructFieldValue("ints", v6),
						)
					}
					v3 = v4
				}
				vs0[i0] = v3
			}
			if len(vs0) == 0 {
				var t1 ydb.Type
				{
					var t2 ydb.Type
					{
						fs0 := make([]ydb.StructOption, 2)
						var t3 ydb.Type
						{
							tp0 := ydb.TypeUTF8
							t3 = ydb.Optional(tp0)
						}
						fs0[0] = ydb.StructField("id", t3)
						var t4 ydb.Type
						{
							var t6 ydb.Type
							{
								tp0 := ydb.TypeInt32
								t6 = ydb.Optional(tp0)
							}
							t5 := ydb.List(t6)
							t4 = t5
						}
						fs0[1] = ydb.StructField("ints", t4)
						t2 = ydb.Struct(fs0...)
					}
					t1 = t2
				}
				t0 := ydb.List(t1)
				list0 = ydb.ZeroValue(t0)
			} else {
				list0 = ydb.ListValue(vs0...)
			}
			v2 = list0
		}
		var v3 ydb.Value
		{
			var list0 ydb.Value
			vs0 := make([]ydb.Value, len(c.Bytes))
			for i0, item0 := range c.Bytes {
				var v4 ydb.Value
				{
					vp0 := ydb.Uint32Value(uint32(item0))
					v4 = vp0
				}
				vs0[i0] = v4
			}
			if len(vs0) == 0 {
				var t1 ydb.Type
				{
					tp0 := ydb.TypeUint32
					t1 = tp0
				}
				t0 := ydb.List(t1)
				list0 = ydb.ZeroValue(t0)
			} else {
				list0 = ydb.ListValue(vs0...)
			}
			v3 = list0
		}
		var v4 ydb.Value
		{
			var list0 ydb.Value
			vs0 := make([]ydb.Value, len(c.Strings))
			for i0, item0 := range c.Strings {
				var v5 ydb.Value
				{
					vp0 := ydb.StringValue([]uint8(item0))
					v5 = vp0
				}
				vs0[i0] = v5
			}
			if len(vs0) == 0 {
				var t1 ydb.Type
				{
					tp0 := ydb.TypeString
					t1 = tp0
				}
				t0 := ydb.List(t1)
				list0 = ydb.ZeroValue(t0)
			} else {
				list0 = ydb.ListValue(vs0...)
			}
			v4 = list0
		}
		var v5 ydb.Value
		{
			vp0 := ydb.OptionalValue(ydb.StringValue(c.String))
			v5 = vp0
		}
		v0 = ydb.StructValue(
			ydb.StructFieldValue("struct", v1),
			ydb.StructFieldValue("structs", v2),
			ydb.StructFieldValue("bytes", v3),
			ydb.StructFieldValue("strings", v4),
			ydb.StructFieldValue("string", v5),
		)
	}
	return v0
}

func (c *Container) StructType() ydb.Type {
	var t0 ydb.Type
	{
		fs0 := make([]ydb.StructOption, 5)
		var t1 ydb.Type
		{
			var t2 ydb.Type
			{
				fs1 := make([]ydb.StructOption, 2)
				var t3 ydb.Type
				{
					tp0 := ydb.TypeUTF8
					t3 = ydb.Optional(tp0)
				}
				fs1[0] = ydb.StructField("id", t3)
				var t4 ydb.Type
				{
					var t6 ydb.Type
					{
						tp0 := ydb.TypeInt32
						t6 = ydb.Optional(tp0)
					}
					t5 := ydb.List(t6)
					t4 = t5
				}
				fs1[1] = ydb.StructField("ints", t4)
				t2 = ydb.Struct(fs1...)
			}
			t1 = t2
		}
		fs0[0] = ydb.StructField("struct", t1)
		var t2 ydb.Type
		{
			var t4 ydb.Type
			{
				var t5 ydb.Type
				{
					fs1 := make([]ydb.StructOption, 2)
					var t6 ydb.Type
					{
						tp0 := ydb.TypeUTF8
						t6 = ydb.Optional(tp0)
					}
					fs1[0] = ydb.StructField("id", t6)
					var t7 ydb.Type
					{
						var t9 ydb.Type
						{
							tp0 := ydb.TypeInt32
							t9 = ydb.Optional(tp0)
						}
						t8 := ydb.List(t9)
						t7 = t8
					}
					fs1[1] = ydb.StructField("ints", t7)
					t5 = ydb.Struct(fs1...)
				}
				t4 = t5
			}
			t3 := ydb.List(t4)
			t2 = t3
		}
		fs0[1] = ydb.StructField("structs", t2)
		var t3 ydb.Type
		{
			var t5 ydb.Type
			{
				tp0 := ydb.TypeUint32
				t5 = tp0
			}
			t4 := ydb.List(t5)
			t3 = t4
		}
		fs0[2] = ydb.StructField("bytes", t3)
		var t4 ydb.Type
		{
			var t6 ydb.Type
			{
				tp0 := ydb.TypeString
				t6 = tp0
			}
			t5 := ydb.List(t6)
			t4 = t5
		}
		fs0[3] = ydb.StructField("strings", t4)
		var t5 ydb.Type
		{
			tp0 := ydb.TypeString
			t5 = ydb.Optional(tp0)
		}
		fs0[4] = ydb.StructField("string", t5)
		t0 = ydb.Struct(fs0...)
	}
	return t0
}

func (f *Foo) Scan(res *table.Result) (err error) {
	res.SeekItem("id")
	f.ID = res.OUTF8()

	res.SeekItem("ints")
	n0 := res.ListIn()
	xs0 := make([]int32, n0)
	for i0 := 0; i0 < n0; i0++ {
		res.ListItem(i0)
		var x0 int32
		x0 = res.OInt32()
		xs0[i0] = x0
	}
	f.Ints = xs0
	res.ListOut()

	return res.Err()
}

func (f *Foo) QueryParameters() *table.QueryParameters {
	var v0 ydb.Value
	{
		vp0 := ydb.OptionalValue(ydb.UTF8Value(f.ID))
		v0 = vp0
	}
	var v1 ydb.Value
	{
		var list0 ydb.Value
		vs0 := make([]ydb.Value, len(f.Ints))
		for i0, item0 := range f.Ints {
			var v2 ydb.Value
			{
				vp0 := ydb.OptionalValue(ydb.Int32Value(item0))
				v2 = vp0
			}
			vs0[i0] = v2
		}
		if len(vs0) == 0 {
			var t1 ydb.Type
			{
				tp0 := ydb.TypeInt32
				t1 = ydb.Optional(tp0)
			}
			t0 := ydb.List(t1)
			list0 = ydb.ZeroValue(t0)
		} else {
			list0 = ydb.ListValue(vs0...)
		}
		v1 = list0
	}
	return table.NewQueryParameters(
		table.ValueParam("$id", v0),
		table.ValueParam("$ints", v1),
	)
}

func (f *Foo) StructValue() ydb.Value {
	var v0 ydb.Value
	{
		var v1 ydb.Value
		{
			vp0 := ydb.OptionalValue(ydb.UTF8Value(f.ID))
			v1 = vp0
		}
		var v2 ydb.Value
		{
			var list0 ydb.Value
			vs0 := make([]ydb.Value, len(f.Ints))
			for i0, item0 := range f.Ints {
				var v3 ydb.Value
				{
					vp0 := ydb.OptionalValue(ydb.Int32Value(item0))
					v3 = vp0
				}
				vs0[i0] = v3
			}
			if len(vs0) == 0 {
				var t1 ydb.Type
				{
					tp0 := ydb.TypeInt32
					t1 = ydb.Optional(tp0)
				}
				t0 := ydb.List(t1)
				list0 = ydb.ZeroValue(t0)
			} else {
				list0 = ydb.ListValue(vs0...)
			}
			v2 = list0
		}
		v0 = ydb.StructValue(
			ydb.StructFieldValue("id", v1),
			ydb.StructFieldValue("ints", v2),
		)
	}
	return v0
}

func (f *Foo) StructType() ydb.Type {
	var t0 ydb.Type
	{
		fs0 := make([]ydb.StructOption, 2)
		var t1 ydb.Type
		{
			tp0 := ydb.TypeUTF8
			t1 = ydb.Optional(tp0)
		}
		fs0[0] = ydb.StructField("id", t1)
		var t2 ydb.Type
		{
			var t4 ydb.Type
			{
				tp0 := ydb.TypeInt32
				t4 = ydb.Optional(tp0)
			}
			t3 := ydb.List(t4)
			t2 = t3
		}
		fs0[1] = ydb.StructField("ints", t2)
		t0 = ydb.Struct(fs0...)
	}
	return t0
}

func (fs *Foos) Scan(res *table.Result) (err error) {
	for res.NextRow() {
		var x0 Foo
		res.SeekItem("id")
		x0.ID = res.OUTF8()

		res.SeekItem("ints")
		n0 := res.ListIn()
		xs0 := make([]int32, n0)
		for i0 := 0; i0 < n0; i0++ {
			res.ListItem(i0)
			var x1 int32
			x1 = res.OInt32()
			xs0[i0] = x1
		}
		x0.Ints = xs0
		res.ListOut()

		if res.Err() == nil {
			*fs = append(*fs, x0)
		}
	}
	return res.Err()
}

func (bs *Bar) Scan(res *table.Result) (err error) {
	for res.NextRow() {
		var x0 [][]string
		if res.Err() == nil {
			*bs = append(*bs, x0)
		}
	}
	return res.Err()
}

func ydbConvU32ToB(x uint32) byte {
	const (
		bits = 8
		mask = (1 << (bits)) - 1
	)
	abs := uint64(x)
	if abs&mask != abs {
		panic(
			"ydbgen: convassert: " + strconv.FormatUint(uint64(x), 10) +
				" (type uint32) overflows byte",
		)
	}
	return byte(x)
}
