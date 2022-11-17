import { styled } from '@mui/styles';
import DashboardLayout from '../layouts/DashboardLayout';
import { useEffect, useState } from 'react';
import {queryPnl, TradeStats} from '../api/bbgo';
import type { BotStats } from '../api/bbgo';

import Detail from '../components/Detail';

const StrategiesContainer = styled('div')(() => ({
  width: '100%',
  height: '100%',
  padding: '40px 20px',
  display: 'grid',

}));

export default function Strategies() {

    const [details, setDetails] = useState<BotStats[]>();
    const [ts, setTs] = useState<TradeStats[]>();

    useEffect(() => {
      queryPnl(  (value) => {
        setDetails(value.report);
        setTs(value.stats);
        console.log(value )
      });

    },[]);

  // @ts-ignore
  return (
    <DashboardLayout>
      <StrategiesContainer>
        {details &&<Detail   data={details} ts={ts} />}
      </StrategiesContainer>
    </DashboardLayout>
  );
}
