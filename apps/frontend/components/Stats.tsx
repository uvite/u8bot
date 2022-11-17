import { styled } from '@mui/styles';
import { StatsTitle, StatsValue, Percentage } from './Summary';
import { BotStats } from '../api/bbgo';

const StatsSection = styled('div')(() => ({
  display: 'grid',
  gridTemplateColumns: '1fr 1fr 1fr',
  gap: '10px',
}));

export default function Stats({
  stats,

}: {
  stats: BotStats;

}) {
  return (
    <StatsSection>
      <div>
        <StatsTitle>Grid Profits</StatsTitle>
        <StatsValue>{stats.profit}</StatsValue>
        <Percentage> {stats.netProfit}</Percentage>
      </div>

      <div>
        <StatsTitle>Floating PNL</StatsTitle>
        <StatsValue>{stats.grossProfit}</StatsValue>
      </div>


    </StatsSection>
  );
}
