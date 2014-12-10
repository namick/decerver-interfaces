package monkjs

import (
	"fmt"
	"github.com/eris-ltd/decerver-interfaces/core"
	"github.com/eris-ltd/decerver-interfaces/events"
	"github.com/eris-ltd/decerver-interfaces/modules"
	"github.com/eris-ltd/thelonious/monk"
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

func (mjs *MonkJs) Restart() error {
	err := mjs.Shutdown()
	if err != nil {
		return nil
	}
	return mjs.Start();
}

func (mjs *MonkJs) SetProperty(name string, data interface{}) {
	
}

func (mjs *MonkJs) Property(name string) interface{} {
	return nil
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
	indata := make([]string, 0)
	for _, d := range data {
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
	return modules.JsReturnVal(nil, nil)
}

func (mjs *MonkJs) AutoCommit(toggle bool) modules.JsObject {
	mjs.mm.AutoCommit(toggle)
	return modules.JsReturnVal(nil, nil)
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
	return modules.JsReturnVal(addresses, nil)
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
	return modules.JsReturnValNoErr(nil)
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

esl.SA = function(acc,addr) {
	return monk.StorageAt(acc,addr).Data;
};

esl.array = {

	//Constants
	"ESizeOffset" : "0",

	"MaxEOffset" : "0",
	"StartOffset" : "1",

	//Structure
	"CTS" : function(name, key){
		return Add(esl.stdvar.Vari(name), Add(Mul(Mod(key, Exp("0x100", "20")), Exp("0x100", "3")), Exp("0x100","2")));
	},
	
	"CTK" : function(slot){
		return Mod(Div(slot, Exp("0x100","3")), Exp("0x100","20"));
	},

	"ESizeslot" : function(name){
		return Add(esl.stdvar.VariBase(name), this.ESizeOffset);
	},
	
	"MaxESlot" : function(name, key){
		return Add(this.CTS(name, key),this.MaxEOffset);
	},
	
	"StartSlot" : function(name, key){
		return Add(this.CTS(name, key),this.StartOffset);
	},

	
	//Gets
	"ESize" : function(addr, name){
		return esl.SA(addr, this.EsizeSlot(name));
	},
	
	"MaxE" : function(addr, name, key){
		return esl.SA(addr, this.MaxESlot(name, key));
	},

	"Element" : function(addr, name, key, index){
		var Esize = this.ESize(addr, name);
		if(this.MaxE(addr, name, key) > index){
			return "0";
		}

		if(Esize == "0x100"){
			return esl.SA(addr, Add(index, this.StartOffset));
		}else{
			var eps = Div("0x100",Esize);
			var pos = Mod(index, eps);
			var row = Add(Mod(Div(index, eps),"0xFFFF"), this.StartOffset);

			var sval = esl.SA(addr, row);
			return Mod(Div(sval, Exp(Esize, pos)), Exp("2", Esize)); 
		}
	},
};

esl.keyvalue = {

	"CTS" : function(name, key){
		return Add(esl.stdvar.Vari(name), Add(Mul(Mod(key, Exp("0x100", "20")), Exp("0x100", "3")), Exp("0x100","2")));
	},
	
	"CTK" : function(slot){
		return Mod(Div(slot, Exp("0x100","3")), Exp("0x100","20"));
	},
	
	"Value" : function(addr, name, key){
		return esl.SA(addr, this.CTS(name, key));
	},
};

esl.ll = {

	//Constants
	"TailSlotOffset"  : "0",
	"HeadSlotOffset"  : "1",
	"LenSlotOffset"   : "2",

	"LLSlotSize" 	  : "3",

	"EntryMainOffset" : "0",
	"EntryPrevOffset" : "1",
	"EntryNextOffset" : "2",

	//Structure
	"CTS" : function(name, key){
		return Add(esl.stdvar.Vari(name), Add(Mul(Mod(key, Exp("0x100", "20")), Exp("0x100", "3")), Exp("0x100","2")));
	},
	
	"CTK" : function(slot){
		return Mod(Div(slot, Exp("0x100","3")), Exp("0x100","20"));
	},

	// Structure
	"TailSlot" : function(name){
		Println("TailSlot");
		return Add(esl.stdvar.VariBase(name), this.TailSlotOffset);
	},
	
	"HeadSlot" : function(name){
		return Add(esl.stdvar.VariBase(name), this.HeadSlotOffset);
	},
	
	"LenSlot" : function(name){
		return Add(esl.stdvar.VariBase(name), this.LenSlotOffset);
	},

	"MainSlot" : function(name, key){
		return Add(this.CTS(name, key), this.EntryMainOffset);
	},
	
	"PrevSlot" : function(name, key){
		return Add(this.CTS(name,key), this.EntryPrevOffset);
	},
	
	"NextSlot" : function(name, key){
		return Add(this.CTS(name,key), this.EntryNextOffset);
	},

	//Gets
	"TailAddr" : function(addr, name){
		var tail = esl.SA(addr, this.TailSlot(name));
		if(IsZero(tail)){
			return null;
		}
		else{
			return tail;
		}
	},
	
	"HeadAddr" : function(addr, name){
		var head = esl.SA(addr, this.HeadSlot(name));
		if(IsZero(head)){
			return null;
		}
		else{
			return head;
		}
	},
	
	"Tail" : function(addr, name){
		var tail = esl.SA(addr, this.TailSlot(name));
		if(IsZero(tail)){
			return null;
		}
		else{
			return this.CTK(tail);
		}
	},
	
	"Head" : function(addr, name){
		var head = esl.SA(addr, this.HeadSlot(name));
		if(IsZero(head)){
			return null;
		}
		else{
			return this.CTK(head);
		}
	},
	
	"Len"  : function(addr, name){
		return esl.SA(addr, this.LenSlot(name));
	},

	"Main" : function(addr, name, key){
		return esl.SA(addr, this.MainSlot(name, key));
	},
	
	"PrevAddr" : function(addr, name, key){
		var prev = esl.SA(addr, this.PrevSlot(name, key));
		if(IsZero(prev)){
			return null;
		}
		else{
			return prev;
		}
	},
	
	"NextAddr" : function(addr, name, key){
		var next = esl.SA(addr, this.NextSlot(name, key));
		if(IsZero(next)){
			return null;
		}
		else{
			return next;
		}
	},
	
	"Prev" : function(addr, name, key){
		var prev = esl.SA(addr, this.PrevSlot(name, key));
		if(IsZero(prev)){
			return null;
		}
		else{
			return this.CTK(prev);
		}	
	},
	
	"Next" : function(addr, name, key){
		var next = esl.SA(addr, this.NextSlot(name, key));
		if(IsZero(next)){
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
			list.push(this.Main(addr, name, current));
			current = this.Next(addr, name, current);
		}

		return list;
	},

	"GetKeys" : function(addr, name){
		var keys = [];
		var current = this.Tail(addr, name);
		while(current !== null){
			list.push(current);
			current = this.Next(addr, name, current);
		}
		return keys;
	},
	
	"GetPairs" : function(addr, name){
       var list = new Array();
       var current = this.Tail(addr, name);
       
       while(current !== null){
           var pair = {};
           pair.Key = current;
           pair.Value = this.Main(addr, name, current);
           list.push(pair);
           current = this.Next(addr, name, current);
       }
       return list;
   },
};

esl.single = {
	
	//Structure
	"ValueSlot" : function(name){
		return esl.stdvar.VariBase(name);
	},
	
	//Gets
	"Value" : function(addr, name){
		return esl.SA(addr, this.ValueSlot(name));
	},
};


esl.double = {
	
	//Structure
	"ValueSlot" : function(name){
		return esl.stdvar.VariBase(name);
	},
	
	"ValueSlot2" : function(name){
		return Add(esl.stdvar.VariBase(name),"1");
	},
	
	//Gets
	"Value" : function(addr, name){
		var values = [];
		Println("Double name: " + name);
		Println("Double Slot: " + this.ValueSlot(name));
		values.push(esl.SA(addr, this.ValueSlot(name)));
		values.push(esl.SA(addr, this.ValueSlot2(name)));
		Println("Double Values: " + values[0]);
		Println("Double Converted: " + HexToString(values[0].slice(2)));
		return values;
	},
};


esl.stdvar = {
	
	//Constants
	"StdVarOffset" 	: "0x1",
	"VarSlotSize" 	: "0x5",
	
	"TypeOffset"	: "0x0",
	"NameOffset"	: "0x1",
	"AddPermOffset"	: "0x2",
	"RmPermOffset"	: "0x3",
	"ModPermOffset"	: "0x4",
	
	//Functions?
	"Vari" 	: function(name){
		var sha3 = SHA3(name);
		var fact = Div(sha3, Exp("0x100", "24") );
		var addr = Add(NSBase, Mul(fact,Exp("0x100", "23")) );
		return addr;
	},
	
	"VarBase" : function(base){
		return Add(base, this.VarSlotSize);
	},
	
	"VariBase" : function(varname){
		return this.VarBase(this.Vari(varname));
	},
	
	//Data Slots
	"VarTypeSlot"	: function(name){
		return Add(this.Vari(name),TypeOffset);
	},
	
	"VarNameSlot"	: function(name){
		return Add(this.Vari(name), NameOffset);
	},
	
	"VarAddPermSlot"	: function(name){
		return Add(this.Vari(name), AddPermOffset);
	},
	
	"VarRmPermSlot" 	: function(name){
		return Add(this.Vari(name), RmPermOffset);
	},
	
	"VarModPermSlot"	: function(name){
		return Add(this.Vari(name), ModPermOffset);
	},
	
	//Getting Variable stuff
	"Type" 	: function(addr, name){
		return esl.SA(addr,this.VarTypeSlot(name));
	},
	
	"Name" 	: function(addr, name){
		return esl.SA(addr,this.VarNameSlot(name));
	},
	
	"Addperm" 	: function(addr, varname){
		return esl.SA(addr,this.VarAddPermSlot(name));
	},
	
	"Rmperm" 	: function(addr, varname){
		return esl.SA(addr,this.VarRmPermSlot(name));
	},
	
	"Modperm" 	: function(addr, varname){
		return esl.SA(addr,this.VarModPermSlot(name));
	},
} 
`
