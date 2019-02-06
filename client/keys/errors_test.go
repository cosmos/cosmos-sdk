package keys

import "testing"

func Test_errKeyNameConflict(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := errKeyNameConflict(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("errKeyNameConflict() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_errMissingName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := errMissingName(); (err != nil) != tt.wantErr {
				t.Errorf("errMissingName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_errMissingPassword(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := errMissingPassword(); (err != nil) != tt.wantErr {
				t.Errorf("errMissingPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_errMissingMnemonic(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := errMissingMnemonic(); (err != nil) != tt.wantErr {
				t.Errorf("errMissingMnemonic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_errInvalidMnemonic(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := errInvalidMnemonic(); (err != nil) != tt.wantErr {
				t.Errorf("errInvalidMnemonic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_errInvalidAccountNumber(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := errInvalidAccountNumber(); (err != nil) != tt.wantErr {
				t.Errorf("errInvalidAccountNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_errInvalidIndexNumber(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := errInvalidIndexNumber(); (err != nil) != tt.wantErr {
				t.Errorf("errInvalidIndexNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
