package monkjs

import (
	"github.com/eris-ltd/decerver-interfaces/core"
	"github.com/eris-ltd/decerver-interfaces/events"
	"github.com/eris-ltd/decerver-interfaces/modules"
	"github.com/eris-ltd/thelonious/monk"
	"fmt"
)

// implements decerver-interfaces Module
type MonkJs struct {
	mm *monk.MonkModule
}

func NewMonkJs() *MonkJs {
	monkModule := monk.NewMonk(nil)
	return &MonkJs{monkModule}
}

// register the module with the decerver javascript vm
func (mjs *MonkJs) Register(fileIO core.FileIO, rm core.RuntimeManager, eReg events.EventRegistry) error {
	rm.RegisterApiObject("monk", mjs)
	rm.RegisterApiScript(eslScript)
	return nil
}

// initialize an monkchain
// it may or may not already have an ethereum instance
// basically gives you a pipe, local keyMang, and reactor
func (mjs *MonkJs) Init() error {
	return mjs.mm.Init()
}

// start the ethereum node
func (mjs *MonkJs) Start() error {
	return mjs.mm.Start()
}

func (mjs *MonkJs) Shutdown() error {
	return mjs.mm.Shutdown()
}

// ReadConfig and WriteConfig implemented in config.go

// What module is this?
func (mjs *MonkJs) Name() string {
	return "monk"
}

func (mjs *MonkJs) Subscribe(name, event, target string) chan events.Event {
	return mjs.mm.Subscribe(name, event, target)
}

func (mjs *MonkJs) UnSubscribe(name string) {
	mjs.mm.UnSubscribe(name)
}

/*
   Wrapper so module satisfies Blockchain
*/

func (mjs *MonkJs) WorldState() modules.JsObject {
	ws := mjs.mm.WorldState()
	return modules.JsReturnVal(modules.ToMap(ws), nil)
}

func (mjs *MonkJs) State() modules.JsObject {
	return modules.JsReturnVal(modules.ToMap(mjs.mm.State()), nil)
}

func (mjs *MonkJs) Storage(target string) modules.JsObject {
	return modules.JsReturnVal(modules.ToMap(mjs.mm.Storage(target)), nil)
}

func (mjs *MonkJs) Account(target string) modules.JsObject {
	return modules.JsReturnVal(modules.ToMap(mjs.mm.Account(target)), nil)
}

func (mjs *MonkJs) StorageAt(target, storage string) modules.JsObject  {
	ret := mjs.mm.StorageAt(target, storage)
	if ret == "" || ret == "0x"{
		ret = "0x0"
	} else {
		ret = "0x" + ret
	}
	
	return modules.JsReturnVal(ret, nil)
}

func (mjs *MonkJs) BlockCount() modules.JsObject {
	return modules.JsReturnVal(mjs.mm.BlockCount(), nil)
}

func (mjs *MonkJs) LatestBlock() modules.JsObject {
	return modules.JsReturnVal(mjs.mm.LatestBlock(), nil)
}

func (mjs *MonkJs) Block(hash string) modules.JsObject {
	return modules.JsReturnVal(modules.ToMap(mjs.mm.Block(hash)), nil)
}

func (mjs *MonkJs) IsScript(target string) modules.JsObject {
	return modules.JsReturnVal(mjs.mm.IsScript(target), nil)
}

func (mjs *MonkJs) Tx(addr, amt string) modules.JsObject {
	hash, err := mjs.mm.Tx(addr, amt)
	var ret modules.JsObject
	if err == nil {
		ret = make(modules.JsObject)
		ret["Hash"] = hash
		ret["Address"] = ""
		ret["Error"] = ""
	}
	return modules.JsReturnVal(ret, err)
}

func (mjs *MonkJs) Msg(addr string, data []interface{}) modules.JsObject {
	indata := make([]string,0)
	for _ , d := range data {
		str, ok := d.(string)
		if !ok {
			return modules.JsReturnValErr(fmt.Errorf("Msg indata is not an array of strings"))		
		}
		indata = append(indata, str)
	}
	hash, err := mjs.mm.Msg(addr, indata)
	var ret modules.JsObject
	if err == nil {
		ret = make(modules.JsObject)
		ret["Hash"] = hash
		ret["Address"] = ""
		ret["Error"] = ""
	}
	return modules.JsReturnVal(ret, err)
}

func (mjs *MonkJs) Script(file, lang string) modules.JsObject {
	addr, err := mjs.mm.Script(file, lang)
	var ret modules.JsObject
	if err == nil {
		ret = make(modules.JsObject)
		ret["Hash"] = ""
		ret["Address"] = addr
		ret["Error"] = ""
	}
	return modules.JsReturnVal(ret, err)
}

func (mjs *MonkJs) Commit() modules.JsObject {
	mjs.mm.Commit()
	return modules.JsReturnVal(nil,nil)
}

func (mjs *MonkJs) AutoCommit(toggle bool) modules.JsObject {
	mjs.mm.AutoCommit(toggle)
	return modules.JsReturnVal(nil,nil)
}

func (mjs *MonkJs) IsAutocommit() modules.JsObject {
	return modules.JsReturnVal(mjs.mm.IsAutocommit(), nil)
}

/*
   Module should also satisfy KeyManager
*/

func (mjs *MonkJs) ActiveAddress() modules.JsObject {
	return modules.JsReturnVal(mjs.mm.ActiveAddress(), nil)
}

func (mjs *MonkJs) Addresses() modules.JsObject {
	count := mjs.mm.AddressCount()
	addresses := make(modules.JsObject)
	array := make([]string, count)
	
	for i := 0; i < count; i++ {
		addr, _ := mjs.mm.Address(i)
		array[i] = addr
	}
	addresses["Addresses"] = array
	return modules.JsReturnVal(addresses,nil)
}

func (mjs *MonkJs) SetAddress(addr string) modules.JsObject {
	err := mjs.mm.SetAddress(addr)
	if err != nil {
		return modules.JsReturnValErr(err)
	} else {
		// No error means success.
		return modules.JsReturnValNoErr(nil)	
	}
}

// TODO Not used atm. Think about this.
func (mjs *MonkJs) SetAddressN(n int) modules.JsObject {
	mjs.mm.SetAddressN(n)
	return modules.JsReturnValNoErr(nil);
}

func (mjs *MonkJs) NewAddress(set bool) modules.JsObject {
	return modules.JsReturnValNoErr(mjs.mm.NewAddress(set))
}

func (mjs *MonkJs) AddressCount() modules.JsObject {
	return modules.JsReturnValNoErr(mjs.mm.AddressCount())
}

var eslScript string = `

var StdVarOffset = "0x1";

var NSBase = Exp("0x100","31");

var esl = {};

GetStorageAt = function(addr,stAddr){
	Println("Getting storage at contract: " + addr + ", address: " + stAddr);
	var rData = monk.StorageAt(addr,stAddr);
	Println("StorageAt Data: " + rData.Data);
	Println("StorageAt Error: " + rData.Error);
	return rData.Data;
}

esl.array = {

	//Structure
	"CTS" : function(name, key){
		return Add(esl.stdvar.Vari(name), Add(Mul(Mod(key, Exp("0x100", "20")), Exp("0x100", "3")), Exp("0x100","2")));
	},
	
	"CTK" : function(slot){
		return Mod(Div(slot, Exp("0x100","3")), Exp("0x100","20"));
	},
	
	"ESizeslot" : function(name){
		return esl.llarray.ESizeslot(esl.stdvar.Vari(name));
	},
	
	"Maxeslot" : function(key){
		return esl.llarray.Maxeslot(this.CTS(name, key));
	},
	
	"StartSlot" : function(key){
		return esl.llarray.StartSlot(this.CTS(name, key));
	},
	
	//Gets
	"ESize" : function(addr, name){
		return esl.llarray.ESize(addr, esl.stdvar.VarBase(esl.stdvar.Vari(name)));
	},
	
	"MaxE" : function(addr, name, key){
		return esl.llarray.MaxE(addr, this.CTS(name, key));
	},
	
	"Element" : function(addr, name, key, index){
		return esl.llarray.Element(addr, esl.stdvar.Vari(name), this.CTS(name, key), index)
	},
};

esl.double = {

	//Structure
	"ValueSlot" : function(varname){
		return Add(stdvar.Vari(varname),stdvar.VarSlotSize);
	},
	
	"ValueSlot2" : function(varname){
		return Add(this.ValueSlot(varname),1);
	},
	
	//Gets
	"Value" : function(addr, varname){
		var values = [];
		values.push(GetStorageAt(addr, this.ValueSlot(varname)).Data);
		values.push(GetStorageAt(addr, this.ValueSlot2(varname)));
		return values;
	},
};

esl.keyvalue = {

	"CTS" : function(name, key){
		return Add(esl.stdvar.Vari(name), Add(Mul(Mod(key, Exp("0x100", "20")), Exp("0x100", "3")), Exp("0x100","2")));
	},
	
	"CTK" : function(slot){
		return Mod(Div(slot, Exp("0x100","3")), Exp("0x100","20"));
	},
	
	"Value" : function(addr, varname, key){
		return esl.llkv.Value(addr, this.CTS(varname, key), "0");
	},
};

esl.ll = {

	//Structure
	"CTS" : function(name, key){
		return Add(esl.stdvar.Vari(name), Add(Mul(Mod(key, Exp("0x100", "20")), Exp("0x100", "3")), Exp("0x100","2")));
	},
	
	"CTK" : function(slot){
		return Mod(Div(slot, Exp("0x100","3")), Exp("0x100","20"));
	},

	"TailSlot" : function(name){
		var addr = esl.stdvar.VariBase(name);
		Println("TailSlot address, move on to esl.lll.TailSlot next: " + addr);
		return esl.llll.TailSlot(addr);
	},
	
	"HeadSlot" : function(name){
	    Println("Headslot");
		return esl.llll.HeadSlot(esl.stdvar.VariBase(name));
	},
	
	"LenSlot" : function(name){
		return esl.llll.LenSlot(esl.stdvar.VariBase(name));
	},

	"MainSlot" : function(name, key){
		return esl.llll.MainSlot(this.CTS(name, key));
	},
	
	"PrevSlot" : function(name, key){
		return esl.llll.Prevlot(this.CTS(name, key));
	},
	
	"NextSlot" : function(name, key){
		return esl.llll.NextSlot(this.CTS(name, key));
	},

	//Gets
	"TailAddr" : function(addr, name){
		var tail = GetStorageAt(addr, this.TailSlot(name));
		if(tail=="0"){
			return null;
		}
		else{
			return tail;
		}
	},
	
	"HeadAddr" : function(addr, name){
		var head = GetStorageAt(addr, this.HeadSlot(name));
		if(head=="0"){
			return null;
		}
		else{
			return head;
		}
	},
	
	"Tail" : function(addr, name){
		var tail = GetStorageAt(addr, this.TailSlot(name));
		Println("Tail function: " + tail);
		if(IsZero(tail)){
			return null;
		}
		else{
			var t = this.CTK(tail);
			Println(t);
			return t;
		}
	},
	
	"Head" : function(addr, name){
		var head = GetStorageAt(addr, this.HeadSlot(name));
		if(head === "0"){
			return null;
		}
		else{
			return this.CTK(head);
		}
	},
	
	"Len"  : function(addr, name){
		return GetStorageAt(addr, this.LenSlot(name));
	},

	"Main" : function(addr, name, key){
		return GetStorageAt(addr, this.MainSlot(name, key));
	},
	
	"PrevAddr" : function(addr, name, key){
		prev=GetStorageAt(addr, this.PrevSlot(name, key));
		if(prev === "0"){
			return null;
		}
		else{
			return prev;
		}
	},
	
	"NextAddr" : function(addr, name, key){
		next=GetStorageAt(addr, this.NextSlot(name, key));
		if(next === "0"){
			return null;
		}
		else{
			return next;
		}
	},
	
	"Prev" : function(addr, name, key){
		prev=GetStorageAt(addr, this.PrevSlot(name, key));
		if(prev === "0"){
			return null;
		}
		else{
			return this.CTK(prev);
		}	
	},
	
	"Next" : function(addr, name, key){
		next=GetStorageAt(addr, this.NextSlot(name, key));
		if(next === "0"){
			return null;
		}
		else{
			return this.CTK(next);
		}
	},

	//Gets the whole list. Note the separate function which gets the keys
	"GetList" : function(addr, name){
		var list = [];
		var current = this.Tail(addr, name);
		while(current !== null){
			list.push(this.Main(addr, current));
			current = this.Next(addr, current);
		}

		return list;
	},

	"GetKeys" : function(addr, name){
		var keys = [];
		var current = this.Tail(addr, name);
		while(current !== null){
			list.push(current);
			current = this.Next(addr, current);
		}

		return keys;
	},

	"GetPairs" : function(addr, name){
	   Println("Getting Pairs");
       var list = [];
       var current = this.Tail(addr, name);
       Println("Current: " + current);
       while(!IsZero(current)){
           var pair = {};
           pair.Key = current;
           pair.Value = this.Main(addr, current);
           list.push(pair);
           current = this.Next(addr, current);
           Println("Current: " + current);
       }
       return list;
   },
};

esl.llarray = {

	//Constants
	"ESizeOffset" : "0",

	"MaxEOffset" : "0",
	"StartOffset" : "1",

	//Structure
	"ESizeslot" : function(base){
		return Add(base, this.ESizeOffset);
	},
	"Maxeslot" : function(slot){
		return Add(slot, this.MaxEOffset);
	},
	"StartSlot" : function(slot){
		return Add(slot, this.StartOffset);
	},

	//Gets
	"ESize" : function(addr, base){
		return GetStorageAt(addr, this.ESizeslot(base));
	},
	
	"MaxE" : function(addr, slot){
		return GetStorageAt(addr, this.Maxeslot(slot));
	},

	"Element" : function(addr, base, slot, index){
		var Esize = this.GetESize(addr, base);
		if(this.GetMaxE(addr, slot) < index){
			return "0";
		}

		if(Esize == "0x100"){
			return GetStorageAt(addr, Add(index, this.StartOffset));
		}else{
			var eps = Div("0x100",Esize);
			var pos = Mod(index, eps);
			var row = Add(Mod(Div(index, eps),"0xFFFF"), this.StartOffset);

			var sval = GetStorageAt(addr, row);
			return Mod(Div(sval, Exp(Esize, pos)), Exp("2", Esize)); 
		}
	},
};

esl.llkv = {

	//Functions
	"Value" : function(addr, slot, offset){
		return GetStorageAt(addr, Add(slot, offset));
	},
};

esl.llll = {

	//Constants
	"TailSlotOffset"  : "0",
	"HeadSlotOffset"  : "1",
	"LenSlotOffset"   : "2",

	"LLLLSlotSize" 	  : "3",

	"EntryMainOffset" : "0",
	"EntryPrevOffset" : "1",
	"EntryNextOffset" : "2",

	// Structure
	"TailSlot" : function(base){
		return Add(base, this.TailSlotOffset);
	},
	
	"HeadSlot" : function(base){
	    Println("Headslot (llll)");
		return Add(base, this.HeadSlotOffset);
	},
	
	"LenSlot" : function(base){
		return Add(base, this.LenSlotOffset);
	},

	"MainSlot" : function(slot){
		return Add(slot, this.EntryMainOffset);
	},
	
	"PrevSlot" : function(slot){
		return Add(slot, this.EntryPrevOffset);
	},
	
	"NextSlot" : function(slot){
		return Add(slot, this.EntryNextOffset);
	},

	//Gets
	"Tail" : function(addr, base){
		return GetStorageAt(addr, this.TailSlot(base));
	},
	
	"Head" : function(addr, base){
		return GetStorageAt(addr, this.HeadSlot(base));
	},
	
	"Len"  : function(addr, base){
		return GetStorageAt(addr, this.LenSlot(base));
	}

	"Main" : function(addr, slot){
		return GetStorageAt(addr, this.MainSlot(slot));
	},
	
	"Prev" : function(addr, slot){
		return GetStorageAt(addr, this.PrevSlot(slot));
	},
	
	"Next" : function(addr, slot){
		return GetStorageAt(addr, this.NextSlot(slot));
	}
};

esl.single = {

	//Structure
	"ValueSlot" : function(varname){
		return esl.stdvar.VarBase(esl.stdvar.Vari(varname));
	},

	//Gets
	"Value" : function(addr, varname){
		return GetStorageAt(addr, this.ValueSlot(varname));
	},
};

esl.stdvar = {

	//Constants
	"VarSlotSize" 	: "0x5",
	"StdVarOffset" 	: "0x1",

	//Functions?
	"Vari" 	: function(varname){
		var sha3 = SHA3(varname);
		Println("Sha3: " + sha3);
		var fact = Div(sha3, Exp("0x100", "24") );
		Println("Fact: " + fact);
		var addr = Add(NSBase, Mul(fact,Exp("0x100", "23")));
		Println("Variable name to address: " + addr);
		return addr;
	},
	
	"VarBase" 	: function(varname){
		return Add(varname, this.StdVarOffset);
	},
	
	"VariBase" : function(varname){
		return this.VarBase(this.Vari(varname))
	},

	//Data Slots
	"VarTypeSlot"	: function(varname){
		return this.Vari(varname);
	},
	
	"VarNameSlot"	: function(varname){
		return Add(this.Vari(varname), 1);
	},
	
	"VarAddPermSlot"	: function(varname){
		return Add(this.Vari(varname), 2);
	},
	
	"VarRmPermSlot" 	: function(varname){
		return Add(this.Vari(varname), 3);
	},
	
	"VarModPermSlot"	: function(varname){
		return Add(this.Vari(varname), 4);
	},

	//Getting Variable stuff
	"Type" 	: function(addr, varname){
		return GetStorageAt(addr,this.VarTypeSlot);
	},
	
	"Name" 	: function(addr, varname){
		return GetStorageAt(addr,this.VarNameSlot);
	},
	
	"Addperm" 	: function(addr, varname){
		return GetStorageAt(addr,this.VarAddPermSlot);
	},
	
	"Rmperm" 	: function(addr, varname){
		return GetStorageAt(addr,this.VarRmPermSlot);
	},
	
	"Modperm" 	: function(addr, varname){
		return GetStorageAt(addr,this.VarModPermSlot);
	},
} 
`
