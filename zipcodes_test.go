package zipcodes

import (
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestNew(t *testing.T) {
	zipcodesDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}
	if (reflect.TypeOf(*zipcodesDataset) != reflect.TypeOf(Zipcodes{})) {
		t.Errorf("Unexpected response type. Got %v, want %v", reflect.TypeOf(*zipcodesDataset), reflect.TypeOf(Zipcodes{}))
	}
}

func TestLoadDataset(t *testing.T) {
	// Wrong file format cases
	cases := []struct {
		Dataset       string
		ExpectedError string
	}{
		{
			"datasets/wrong_length_dataset.txt",
			"zipcodes: file line does not have 12 fields",
		},
		{
			"datasets/wrong_lat_dataset.txt",
			"zipcodes: error while converting WRONG to Latitude",
		},
		{
			"datasets/wrong_lon_dataset.txt",
			"zipcodes: error while converting WRONG to Longitude",
		},
	}

	for _, c := range cases {
		_, err := LoadDataset(c.Dataset)
		if err.Error() != c.ExpectedError {
			t.Errorf("Unexpected error. Got %s, want %s", err, c.ExpectedError)
		}
	}

	// Valid file format cases
	dataset, err := LoadDataset("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}
	if (reflect.TypeOf(dataset) != reflect.TypeOf(Zipcodes{})) {
		t.Errorf("Unexpected response type. Got %v, want %v", reflect.TypeOf(dataset), reflect.TypeOf(Zipcodes{}))
	}
}

func TestLoadDatasetReader(t *testing.T) {
	t.Run("Wrong file format cases", func(t *testing.T) {
		cases := []struct {
			Dataset       string
			ExpectedError string
		}{
			{
				"datasets/wrong_length_dataset.txt",
				"zipcodes: file line does not have 12 fields",
			},
			{
				"datasets/wrong_lat_dataset.txt",
				"zipcodes: error while converting WRONG to Latitude",
			},
			{
				"datasets/wrong_lon_dataset.txt",
				"zipcodes: error while converting WRONG to Longitude",
			},
		}

		for _, c := range cases {
			file, err := os.Open(c.Dataset)
			if err != nil {
				t.Errorf("Unexpected error while opening dataset %v", err)
			}
			defer file.Close()
			_, err = LoadDatasetReader(file)
			if err.Error() != c.ExpectedError {
				t.Errorf("Unexpected error. Got %s, want %s", err, c.ExpectedError)
			}
		}
	})

	// unexpected readers
	t.Run("unexpected readers", func(t *testing.T) {
		cases := []struct {
			Reader        io.Reader
			ExpectedError string
		}{
			{
				nil,
				"zipcodes: unexpected nil reader",
			},
			{
				strings.NewReader("invalid format"),
				"zipcodes: file line does not have 12 fields",
			},
		}
		for _, c := range cases {
			_, err := LoadDatasetReader(c.Reader)
			if err == nil {
				t.Errorf("Expected error, got nil")
				continue
			}
			if err.Error() != c.ExpectedError {
				t.Errorf("Unexpected error. Got %s, want %s", err, c.ExpectedError)
			}
		}
	})

	// Valid file format cases
	t.Run("valid readers - valid_dataset.txt", func(t *testing.T) {
		file, err := os.Open("datasets/valid_dataset.txt")
		if err != nil {
			t.Errorf("Unexpected error while opening dataset %v", err)
		}
		defer file.Close()
		dataset, err := LoadDatasetReader(file)
		if err != nil {
			t.Errorf("Unexpected error while initializing struct %v", err)
		}
		if (reflect.TypeOf(dataset) != reflect.TypeOf(Zipcodes{})) {
			t.Errorf("Unexpected response type. Got %v, want %v", reflect.TypeOf(dataset), reflect.TypeOf(Zipcodes{}))
		}
	})
	t.Run("valid readers - empty reader", func(t *testing.T) {
		dataset, err := LoadDatasetReader(strings.NewReader(""))
		if err != nil {
			t.Errorf("Unexpected error while initializing struct %v", err)
		}
		if (reflect.TypeOf(dataset) != reflect.TypeOf(Zipcodes{})) {
			t.Errorf("Unexpected response type. Got %v, want %v", reflect.TypeOf(dataset), reflect.TypeOf(Zipcodes{}))
		}
	})
}

func TestLookup(t *testing.T) {
	zipcodesDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}

	// Looking for a zipcode that exists
	existingZipCode := "01945"
	foundedZC, err := zipcodesDataset.Lookup(existingZipCode)
	if err != nil {
		t.Errorf("Unexpected error while looking for zipcode %s", existingZipCode)
	}
	expectedZipCode := ZipCodeLocation{
		ZipCode:   "01945",
		PlaceName: "Guteborn",
		AdminName: "Brandenburg",
		Lat:       51.4167,
		Lon:       13.9333,
	}

	if reflect.DeepEqual(foundedZC, &expectedZipCode) != true {
		t.Errorf("Unexpected response when calling Lookup")
	}
	// Looking for a zipcode that does not exists
	missingZipCode := "XYZ"
	_, errZC := zipcodesDataset.Lookup(missingZipCode)
	if errZC.Error() != "zipcodes: zipcode XYZ not found !" {
		t.Errorf("Unexpected error while looking for zipcode %s", existingZipCode)
	}
}

func TestDistanceBetweenPoints(t *testing.T) {
	cases := []struct {
		coordsA    []float64
		coordsB    []float64
		ExpectedKM float64
	}{
		{
			[]float64{52.520008, 13.404954}, // Berlin
			[]float64{51.217941, 6.761680},  // Düsseldorf
			478.34,
		},
		{
			[]float64{40.730610, -73.935242}, // New York
			[]float64{40.416775, -3.703790},  // Madrid
			5761.76,
		},
		{
			[]float64{13.736717, 100.523186}, // Bangkok
			[]float64{18.796143, 98.979263},  // Chiang Mai
			586.21,
		},
	}

	for _, c := range cases {
		kms := DistanceBetweenPoints(c.coordsA[0], c.coordsA[1], c.coordsB[0], c.coordsB[1], earthRadiusKm)
		if kms != c.ExpectedKM {
			t.Errorf("Distance does not match. Expected %v, got %v", c.ExpectedKM, kms)
		}
	}
}

func TestCalculateDistance(t *testing.T) {
	// Testing valid cases where the postal code exists
	cases := []struct {
		ZipCodeA   string
		ZipCodeB   string
		ExpectedKM float64
	}{
		{
			"01945",
			"03058",
			49.87,
		},
		{
			"20457",
			"22525",
			7.43,
		},
		{
			"19053",
			"87787",
			643.03,
		},
	}

	zipcodesDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}

	for _, c := range cases {
		kms, err := zipcodesDataset.CalculateDistance(c.ZipCodeA, c.ZipCodeB, earthRadiusKm)
		if err != nil {
			t.Errorf("Unexpected error while looking for zipcode %s", err)
		}
		if kms != c.ExpectedKM {
			t.Errorf("Distance does not match. Expected %v, got %v", c.ExpectedKM, kms)
		}
	}

	// Testing cases where the postal code does not exists
	fail := []struct {
		ZipCodeA    string
		ZipCodeB    string
		ExpectedErr string
	}{
		{
			"01945",
			"11111",
			"zipcodes: zipcode 11111 not found !",
		},
		{
			"00000",
			"22525",
			"zipcodes: zipcode 00000 not found !",
		},
	}

	zcDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}

	for _, c := range fail {
		_, err := zcDataset.CalculateDistance(c.ZipCodeA, c.ZipCodeB, earthRadiusKm)
		if err.Error() != c.ExpectedErr {
			t.Errorf("Unexpected error. Got %s, want %s", err, c.ExpectedErr)
		}
	}
}

func TestDistanceInKm(t *testing.T) {
	cases := []struct {
		ZipCodeA   string
		ZipCodeB   string
		ExpectedKM float64
	}{
		{
			"01945",
			"03058",
			49.87,
		},
		{
			"20457",
			"22525",
			7.43,
		},
		{
			"19053",
			"87787",
			643.03,
		},
	}

	zipcodesDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}

	for _, c := range cases {
		kms, err := zipcodesDataset.DistanceInKm(c.ZipCodeA, c.ZipCodeB)
		if err != nil {
			t.Errorf("Unexpected error while looking for zipcode %s", err)
		}
		if kms != c.ExpectedKM {
			t.Errorf("Distance does not match. Expected %v, got %v", c.ExpectedKM, kms)
		}
	}
}

func TestDistanceInMiles(t *testing.T) {
	cases := []struct {
		ZipCodeA   string
		ZipCodeB   string
		ExpectedMi float64
	}{
		{
			"01945",
			"03058",
			30.98,
		},
		{
			"20457",
			"22525",
			4.62,
		},
		{
			"19053",
			"87787",
			399.48,
		},
	}

	zipcodesDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}

	for _, c := range cases {
		miles, err := zipcodesDataset.DistanceInMiles(c.ZipCodeA, c.ZipCodeB)
		if err != nil {
			t.Errorf("Unexpected error while looking for zipcode %s", err)
		}
		if miles != c.ExpectedMi {
			t.Errorf("Distance does not match. Expected %v, got %v", c.ExpectedMi, miles)
		}
	}
}

func TestDistanceInKmToZipCode(t *testing.T) {
	cases := []struct {
		ZipCode          string
		Latitude         float64
		Longitude        float64
		ExpectedResponse float64
	}{
		{
			"01945",
			51.4267,
			13.9333,
			1.11,
		},
		{
			"01945",
			51.4067,
			13.9333,
			1.11,
		},
	}

	zipcodesDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}

	for _, c := range cases {
		kms, err := zipcodesDataset.DistanceInKmToZipCode(c.ZipCode, c.Latitude, c.Longitude)

		if err != nil {
			t.Errorf("Unexpected error while looking for zipcode %s", err)
		}

		if kms != c.ExpectedResponse {
			t.Errorf("Expected distance in kilometers to zipcode does not match. Expected %v, got %v", c.ExpectedResponse, kms)
		}
	}
}

func TestDistanceInMilToZipCode(t *testing.T) {
	cases := []struct {
		ZipCode          string
		Latitude         float64
		Longitude        float64
		ExpectedResponse float64
	}{
		{
			"01945",
			51.4267,
			13.9333,
			0.69,
		},
		{
			"01945",
			51.4067,
			13.9333,
			0.69,
		},
	}

	zipcodesDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}

	for _, c := range cases {
		miles, err := zipcodesDataset.DistanceInMilToZipCode(c.ZipCode, c.Latitude, c.Longitude)

		if err != nil {
			t.Errorf("Unexpected error while looking for zipcode %s", err)
		}

		if miles != c.ExpectedResponse {
			t.Errorf("Expected distance in miles to zipcode does not match. Expected %v, got %v", c.ExpectedResponse, miles)
		}
	}
}

func TestGetZipcodesWithinKmRadius(t *testing.T) {
	cases := []struct {
		ZipCode          string
		Radius           float64
		ExpectedResponse []string
	}{
		{
			"01945",
			50.0,
			[]string{"03058"},
		},
		{
			"01945",
			100.0,
			[]string{"03058"},
		},
	}
	zipcodesDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}
	for _, c := range cases {
		zcList, err := zipcodesDataset.GetZipcodesWithinKmRadius(c.ZipCode, c.Radius)
		if err != nil {
			t.Errorf("Unexpected error while looking for zipcode %s", err)
		}

		if reflect.DeepEqual(zcList, c.ExpectedResponse) != true {
			t.Errorf("Unxpected zipcode list returned.")
		}
	}
}

func TestGetZipcodesWithinMlRadius(t *testing.T) {
	cases := []struct {
		ZipCode          string
		Radius           float64
		ExpectedResponse []string
	}{
		{
			"01945",
			50.0,
			[]string{"03058"},
		},
		{
			"01945",
			100.0,
			[]string{"03058"},
		},
	}
	zipcodesDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}
	for _, c := range cases {
		zcList, err := zipcodesDataset.GetZipcodesWithinMlRadius(c.ZipCode, c.Radius)
		if err != nil {
			t.Errorf("Unexpected error while looking for zipcode %s", err)
		}

		if reflect.DeepEqual(zcList, c.ExpectedResponse) != true {
			t.Errorf("Unxpected zipcode list returned.")
		}
	}
}

func TestFindZipcodesWithinRadius(t *testing.T) {
	cases := []struct {
		Location     *ZipCodeLocation
		MaxRadius    float64
		EarthRadius  float64
		ExpectedList []string
	}{
		{
			&ZipCodeLocation{
				ZipCode:   "01945",
				PlaceName: "Guteborn",
				AdminName: "Brandenburg",
				Lat:       51.4167,
				Lon:       13.9333,
			},
			50,
			earthRadiusKm,
			[]string{"03058"},
		},
		{
			&ZipCodeLocation{
				ZipCode:   "01945",
				PlaceName: "Guteborn",
				AdminName: "Brandenburg",
				Lat:       51.4167,
				Lon:       13.9333,
			},
			50,
			earthRadiusKm,
			[]string{"03058"},
		},
	}
	zipcodesDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}
	for _, c := range cases {
		list := zipcodesDataset.FindZipcodesWithinRadius(c.Location, c.MaxRadius, c.EarthRadius)

		if reflect.DeepEqual(list, c.ExpectedList) != true {
			t.Errorf("FindZipcodesWithinRadius returned an unexpected zipcode list.")
		}
	}
}

func TestLookupByCityState(t *testing.T) {
	zipcodesDataset, err := New("datasets/valid_dataset.txt")
	if err != nil {
		t.Errorf("Unexpected error while initializing struct %v", err)
	}

	spew.Config.Indent = "\t"
	spew.Dump(`zipcodesDataset: %#v\n`, zipcodesDataset)
	found := zipcodesDataset.LookupByCityState("Hamburg Neustadt", "Hamburg")
	if len(found) == 0 {
		t.Error("not found for Hamburg Neustadt Hamburg")
	}

	found = zipcodesDataset.LookupByCityState("Hamburg Neustadt", "HH")
	if len(found) == 0 {
		t.Error("not found for Hamburg Neustadt HH")
	}

	found = zipcodesDataset.LookupByCityState("hamburg neustadt", "hamburg")
	if len(found) == 0 {
		t.Error("not found for hamburg neustadt hamburg")
	}

	found = zipcodesDataset.LookupByCityState("hamburg neustadt", "hh")
	if len(found) == 0 {
		t.Error("not found for hamburg neustadt hh")
	}

	found = zipcodesDataset.LookupByCityState("something", "hh")
	if len(found) != 0 {
		t.Error("found for something hh")
	}

	found = zipcodesDataset.LookupByCityState("hamburg neustadt", "something")
	if len(found) != 0 {
		t.Error("found for hamburg neustadt something")
	}
}
