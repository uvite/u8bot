import {styled} from '@mui/styles';
import type {BotStats, TradeStats} from '../api/bbgo';
import { getCookie} from '../api/cooke';

import RunningTime from './RunningTime';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';

function createData(
    key: string,
    value: any,
) {
    return {key, value};
}



const StrategyContainer = styled('section')(() => ({
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'space-around',
    width: '100%',
    border: '1px solid rgb(248, 149, 35)',
    borderRadius: '10px',
    padding: '10px',
}));

const Strategy = styled('div')(() => ({
    fontSize: '20px',
}));

export const Description = styled('div')(() => ({
    color: 'rgb(140, 140, 140)',
    '& .duration': {
        marginLeft: '3px',
    },
}));

export default function Detail({data,ts}: { data: BotStats,ts:TradeStats }) {
    const {startTime} = data;
    const strategy=getCookie("strategy")
    const Status=getCookie("Status")
    const symbol=getCookie("symbol")
    // const totalProfitsPercentage = (stats.totalProfits / stats.investment) * 100;
    // const gridProfitsPercentage = (stats.gridProfits / stats.investment) * 100;
    // const gridAprPercentage = (stats.gridProfits / 5) * 365;

    const now = Date.now();
    const durationMilliseconds = now - Date.parse(startTime.toString());
    const seconds = durationMilliseconds / 1000;
    // @ts-ignore
    const rows = [
        createData("交易数量",data.numTrades),
        createData("买入量",data.buyVolume),
        createData("卖出量",data.sellVolume),
        createData("未实现赢利",data.unrealizedProfit),
        createData("净利润（Net Profit）",data.netProfit),
        createData("毛利润（Gross Profit",data.grossProfit),
        createData("毛亏损（Gross Loss",data.grossLoss),
        createData("最大交易亏损 (Max Drawdown)",data.profit),
        createData("买入和持有收益 (Buy & Hold Return)",data.profit),
        createData("夏普比率 (Sharpe Ratio)",data.profit),
        createData("索提诺比率(Sortino Ratio)",data.profit),
        createData("盈利因子(Profit Factor)",data.profit),
        createData("最大持仓 (Max Contracts Held)",data.profit),
        createData("佣金 (Commission Paid)",data.feeInUSD),
        createData("所有已平仓交易 (Total Closed Trades)",data.profit),
        createData("总持仓 (Total Open Trades)",data.profit),
        createData("获利交易次数 (Number Winning Trades)",ts.numOfLossTrade),
        createData("亏损交易次数 (Number Losing Trades)",ts.numOfProfitTrade),
        createData("胜率 (Percent Profitable)",ts.winningRatio),
        createData("平均交易 (Avg Trade)",data.profit),
        createData("平均盈利交易 (Avg Win Trade)",ts.averageProfitTrade),
        createData("平均亏损交易(Avg Loss Trade)",ts.averageLossTrade),
        createData("平均盈利/平均亏损 (Ratio Avg Win / Avg Loss)",ts.averageProfitTrade),
        createData("最大盈利交易 (Largest Win Trade)",ts.largestProfitTrade),
        createData("最大亏损交易 (Largest Losing Trade)",ts.largestLossTrade),
        createData("全部交易的平均持仓K线根数 (Avg # Bars in Trades)",data.profit),
        createData("盈利交易的平均持仓K线根数 (Avg # Bars in Winning Trades)",data.profit),
        createData("亏损交易的平均持仓K线根数 (Avg # Bars in Losing Trades)",data.profit),
        createData("全部交易的平均持仓K线根数 (Avg # Bars in Trades)",data.profit),

        createData("最大连续赢利 次数  ",ts.maximumConsecutiveWins),
        createData("最大连续亏损 次数",ts.maximumConsecutiveLosses),
        createData("最大连续利润  ",ts.maximumConsecutiveProfit),
        createData("最大连续亏损  ",ts.maximumConsecutiveLoss),

    ];

    // 最大连续赢利 （$） ― 最长的获胜交易系列及其总利润;
    // 最大连续亏损 （$） ― 最长的亏损交易系列及其总亏损;
    // 最大连续利润（计数） ― 一系列盈利交易的最大利润和该序列中盈利交易的数量;
    // 最大连续亏损（计数） ― 一系列亏损交易的最大亏损及其亏损交易的数量;
    //     profitFactor: 2.63171056,
    //     totalNetProfit: 2.67359636,
    //     maximumConsecutiveWins: 2,
    //     maximumConsecutiveLosses: 5,
    //     maximumConsecutiveProfit: 1.98999998,
    //     maximumConsecutiveLoss: -0.70682374

    return (
        <StrategyContainer>
            <Strategy>{strategy} 状态{Status}</Strategy>
            <div>{data.symbol}</div>
            <RunningTime seconds={seconds}/>
            <Description>
                策略说明
            </Description>

            <TableContainer component={Paper}>
                <Table sx={{minWidth: 650}} aria-label="simple table">

                    <TableBody>
                        {rows.map((row) => (
                            <TableRow
                                key={row.key}
                                sx={{'&:last-child td, &:last-child th': {border: 0}}}
                            >
                                <TableCell component="th" scope="row">
                                    {row.key}
                                </TableCell>
                                <TableCell align="right">{row.value}</TableCell>

                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </TableContainer>
        </StrategyContainer>
    );
}
