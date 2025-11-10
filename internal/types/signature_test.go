package types

import (
	"math/big"
	"reflect"
	"testing"
)

func TestSignature_V(t *testing.T) {
	type fields struct {
		R          *big.Int
		S          *big.Int
		OddYParity bool
	}
	type args struct {
		chainID  *big.Int
		isLegacy bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *big.Int
	}{
		{
			name: "legacy non-eip155 even parity",
			fields: fields{
				R:          big.NewInt(1),
				S:          big.NewInt(2),
				OddYParity: false,
			},
			args: args{
				chainID:  nil,
				isLegacy: true,
			},
			want: big.NewInt(27),
		},
		{
			name: "legacy non-eip155 odd parity",
			fields: fields{
				R:          big.NewInt(1),
				S:          big.NewInt(2),
				OddYParity: true,
			},
			args: args{
				chainID:  nil,
				isLegacy: true,
			},
			want: big.NewInt(28),
		},
		{
			name: "legacy eip155 even parity",
			fields: fields{
				R:          big.NewInt(1),
				S:          big.NewInt(2),
				OddYParity: false,
			},
			args: args{
				chainID:  big.NewInt(100),
				isLegacy: true,
			},
			want: big.NewInt(235),
		},
		{
			name: "legacy eip155 odd parity",
			fields: fields{
				R:          big.NewInt(1),
				S:          big.NewInt(2),
				OddYParity: true,
			},
			args: args{
				chainID:  big.NewInt(100),
				isLegacy: true,
			},
			want: big.NewInt(236),
		},
		{
			name: "dynamic fee even parity",
			fields: fields{
				R:          big.NewInt(1),
				S:          big.NewInt(2),
				OddYParity: false,
			},
			args: args{
				chainID:  big.NewInt(100),
				isLegacy: false,
			},
			want: big.NewInt(0),
		},
		{
			name: "dynamic fee odd parity",
			fields: fields{
				R:          big.NewInt(1),
				S:          big.NewInt(2),
				OddYParity: true,
			},
			args: args{
				chainID:  big.NewInt(100),
				isLegacy: false,
			},
			want: big.NewInt(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signature{
				R:          tt.fields.R,
				S:          tt.fields.S,
				OddYParity: tt.fields.OddYParity,
			}
			if got := s.V(tt.args.chainID, tt.args.isLegacy); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Signature.V() = %v, want %v", got, tt.want)
			}
		})
	}
}
