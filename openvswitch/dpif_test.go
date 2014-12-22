package openvswitch

import (
	"testing"
	"math/rand"
	"fmt"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func checkedCloseDpif(dpif *Dpif, t *testing.T) {
	err := dpif.Close()
	if err != nil { t.Fatal(err) }
}

func TestCreateDatapath(t *testing.T) {
	dpif, err := NewDpif()
	if err != nil { t.Fatal(err) }
	defer checkedCloseDpif(dpif, t)

	name := fmt.Sprintf("test%d", rand.Intn(100000))

	dp, err := dpif.CreateDatapath(name)
	if err != nil { t.Fatal(err) }

	err = dp.Delete()
	if err != nil { t.Fatal(err) }
}

func TestLookupDatapath(t *testing.T) {
	dpif, err := NewDpif()
	if err != nil { t.Fatal(err) }
	defer checkedCloseDpif(dpif, t)

	name := fmt.Sprintf("test%d", rand.Intn(100000))
	dp, err := dpif.LookupDatapath(name)
	if err != nil { t.Fatal(err) }
	if dp != nil { t.Fatal("LookupDatapath should return nil for non-existent name") }

	_, err = dpif.CreateDatapath(name)
	if err != nil { t.Fatal(err) }

	checkedCloseDpif(dpif, t)
	dpif, err = NewDpif()
	if err != nil { t.Fatal(err) }
	defer checkedCloseDpif(dpif, t)

	dp, err = dpif.LookupDatapath(name)
	if err != nil { t.Fatal(err) }
	if dp == nil { t.Fatal("LookupDatapath returned nil") }

	err = dp.Delete()
	if err != nil { t.Fatal(err) }
}

func TestEnumerateDatapaths(t *testing.T) {
	dpif, err := NewDpif()
	if err != nil { t.Fatal(err) }
	defer checkedCloseDpif(dpif, t)

	const n = 10
	var names [n]string
	var dps [n]*Datapath

	cleanup := func () {
		for i, dp := range(dps) {
			if dp != nil {
				dp.Delete()
			}

			dps[i] = nil
		}
	}

	defer cleanup()

	for i := range(names) {
		names[i] = fmt.Sprintf("test%d", rand.Intn(100000))
		dp, err := dpif.CreateDatapath(names[i])
		if err != nil { t.Fatal(err) }
		dps[i] = dp
	}

	name2dp, err := dpif.EnumerateDatapaths()
	if err != nil { t.Fatal(err) }
	for _, name := range(names) {
		_, ok := name2dp[name]
		if !ok { t.Fatal() }
	}

	cleanup()

	name2dp, err = dpif.EnumerateDatapaths()
	if err != nil { t.Fatal(err) }
	for _, name := range(names) {
		_, ok := name2dp[name]
		if ok { t.Fatal() }
	}
}

func checkedDeleteDatapath(dp *Datapath, t *testing.T) {
	err := dp.Delete()
	if err != nil { t.Fatal(err) }
}

func TestCreateVport(t *testing.T) {
	dpif, err := NewDpif()
	if err != nil { t.Fatal(err) }
	defer checkedCloseDpif(dpif, t)

	dp, err := dpif.CreateDatapath(fmt.Sprintf("test%d", rand.Intn(100000)))
	if err != nil { t.Fatal(err) }
	defer checkedDeleteDatapath(dp, t)

	name := fmt.Sprintf("test%d", rand.Intn(100000))
	vport, err := dp.CreateVport(name, INTERNAL_VPORT_SPEC)
	if err != nil { t.Fatal(err) }

	err = vport.Delete()
	if err != nil { t.Fatal(err) }
}

func TestLookupVport(t *testing.T) {
	dpif, err := NewDpif()
	if err != nil { t.Fatal(err) }
	defer checkedCloseDpif(dpif, t)

	dpname := fmt.Sprintf("test%d", rand.Intn(100000))
	dp, err := dpif.CreateDatapath(dpname)
	if err != nil { t.Fatal(err) }

	name := fmt.Sprintf("test%d", rand.Intn(100000))
	vport, err := dp.LookupVport(name)
	if !IsNoSuchVportError(err) { t.Fatal(err) }

	_, err = dp.CreateVport(name, INTERNAL_VPORT_SPEC)
	if err != nil { t.Fatal(err) }

	checkedCloseDpif(dpif, t)
	dpif, err = NewDpif()
	if err != nil { t.Fatal(err) }
	defer checkedCloseDpif(dpif, t)

	dp, err = dpif.LookupDatapath(dpname)
	if err != nil { t.Fatal(err) }
	if dp == nil { t.Fatal("LookupDatapath returned nil") }
	defer dp.Delete()

	vport, err = dp.LookupVport(name)
	if err != nil { t.Fatal(err) }

	err = vport.Delete()
	if err != nil { t.Fatal(err) }
}

func TestEnumerateVports(t *testing.T) {
	dpif, err := NewDpif()
	if err != nil { t.Fatal(err) }
	defer checkedCloseDpif(dpif, t)

	dp, err := dpif.CreateDatapath(fmt.Sprintf("test%d", rand.Intn(100000)))
	if err != nil { t.Fatal(err) }
	defer checkedDeleteDatapath(dp, t)

	var names []string
	var vports []VportHandle

	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("test%d", rand.Intn(100000))
		vport, err := dp.CreateVport(name, INTERNAL_VPORT_SPEC)
		if err != nil { t.Fatal(err) }
		names = append(names, name)
		vports = append(vports, vport)
	}

	name2vport, err := dp.EnumerateVports()
	if err != nil { t.Fatal(err) }
	for _, name := range(names) {
		_, ok := name2vport[name]
		if !ok { t.Fatal() }
	}

	for _, vport := range(vports) {
		vport.Delete()
	}

	name2vport, err = dp.EnumerateVports()
	if err != nil { t.Fatal(err) }
	for _, name := range(names) {
		_, ok := name2vport[name]
		if ok { t.Fatal() }
	}
}

func TestCreateFlow(t *testing.T) {
	dpif, err := NewDpif()
	if err != nil {	t.Fatal(err) }
	defer checkedCloseDpif(dpif, t)

	dp, err := dpif.CreateDatapath(fmt.Sprintf("test%d", rand.Intn(100000)))
	if err != nil {	t.Fatal(err) }
	defer checkedDeleteDatapath(dp, t)

	f := NewFlowSpec()
	f.AddKey(NewEthernetFlowKey([...]byte { 1,2,3,4,5,6 }, [...]byte { 6,5,4,3,2,1 }))

	err = dp.CreateFlow(f)
	if err != nil {	t.Fatal(err) }

	err = dp.DeleteFlow(f)
	if err != nil {	t.Fatal(err) }

	err = dp.DeleteFlow(f)
	if err != (NoSuchFlowError{}) {
		t.Fatal()
	}
}

func TestEnumerateFlows(t *testing.T) {
	dpif, err := NewDpif()
	if err != nil { t.Fatal(err) }
	defer checkedCloseDpif(dpif, t)

	dp, err := dpif.CreateDatapath(fmt.Sprintf("test%d", rand.Intn(100000)))
	if err != nil { t.Fatal(err) }
	defer checkedDeleteDatapath(dp, t)

	const n = 10
	var flows [n]FlowSpec

	for i := range(flows) {
		flow := NewFlowSpec()
		flow.AddKey(NewEthernetFlowKey([...]byte { 1,2,3,4,5,byte(i) }, [...]byte { 6,5,4,3,2,1 }))
		err = dp.CreateFlow(flow)
		if err != nil { t.Fatal(err) }
		flows[i] = flow
	}

	eflows, err := dp.EnumerateFlows()
	if err != nil { t.Fatal(err) }

	if len(eflows) != n { t.Fatal() }

	for _, eflow := range(eflows) {
		found := false

		for _, flow := range(flows) {
			if eflow.Equals(flow) {
				found = true
				break
			}
		}

		if !found { t.Fatal(eflow) }
	}

	for _, eflow := range(eflows) {
		err = dp.DeleteFlow(eflow)
		if err != nil { t.Fatal(err) }
	}

	eflows, err = dp.EnumerateFlows()
	if err != nil { t.Fatal(err) }

	if len(eflows) != 0 { t.Fatal() }
}