import * as echarts from 'echarts/core';

import { BarChart, LineChart } from 'echarts/charts';

import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
} from 'echarts/components';

import { SVGRenderer } from 'echarts/renderers';

// 按需注册：当前仅控制台柱状图与 Sprint 燃尽折线图在用。
// 需要新图表类型（饼图/雷达/地图/缩放等）时在此追加对应模块，
// 全量注册曾使 chunk 达 778KB
echarts.use([
  BarChart,
  LineChart,
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  SVGRenderer,
]);

export default echarts;
