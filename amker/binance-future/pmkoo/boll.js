
let bb=ta.boll("SMA",close,20,2,2)

console.log("=========")

console.log("close",close.tail(3).reverse())


console.log("u",bb.u.tail(3))
console.log("m",bb.m.tail(3))
console.log("l",bb.l.tail(3))


console.log("=========\n")
