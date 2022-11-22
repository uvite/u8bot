


console.log(close)
let abc=ta.rsi(close,5)
let aa=ta.series()
aa.push("hold")
aa.push("buy")
aa.push("buy")
aa.push("sell")
aa.push("sell")
let changed=ta.change(aa,2)
console.log(abc.tail(5),changed)
// console.log(close.tail(5),high.tail(5),low.last(5))
// //console.log(high.last())
// let c=ta.atr(high,low,close,3)
//
// console.log(c)
// console.log(c.tail(5))
// let sma=ta.sma(close,30)
// console.log(sma.tail(5))

console.log("==jsend===")
