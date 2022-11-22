

export function init(){

    console.log(4444)
}

export function shouldOpen(){
    return true
}

export function shouldShort(){
    return false
}

export function onKlineClose(kline){
    console.log(kline)
}