
// let aa=ta.series()
// aa.push("hold")
// aa.push("buy")
// aa.push("buy")
// aa.push("sell")
// aa.push("sell")
// let changed=ta.change(aa,2)

let rsi=ta.rsi(close,14)
let sma=ta.sma(close,14)
let atr=ta.atr(high,low,close,14)
let ema=ta.ema(close,14)
let cci=ta.cci(high,low,close,14)
let kama=ta.kama(close,14)
let wma=ta.wma(close,14)

let hma=ta.hma(close,14)


//alma 暂时没通过
let alma=ta.alma(close,14,0.85,6)


console.log("=========")

console.log("close",close.tail(3).reverse())

console.log("rsi",rsi.tail(3).reverse() )
console.log("sma",sma.tail(3).reverse() )
console.log("atr",atr.tail(3).reverse())
console.log("ema",ema.tail(3).reverse())
console.log("cci",cci.tail(3).reverse())
console.log("kama",kama.tail(3).reverse())
console.log("wma",wma.tail(3).reverse())
console.log("alma",kama.tail(3).reverse())
console.log("hma",hma.tail(3).reverse())

//自定义
let jma=function(){

    console.log("function jma",close.last())
}
jma()
//console.log(high.last())

// let c=ta.sma(fuck,10)

console.log("=========\n")
