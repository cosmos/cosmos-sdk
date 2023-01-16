import * as d3 from 'd3';
import { ColorMap } from 'src/types';

type divEl = d3.Selection<HTMLDivElement, unknown, null, undefined>;

export const legend = (
    target: divEl,
    names: string[],
    transformName: Function,
    colorMap: ColorMap,
    onClick: Function,
    nameToHighlight: string | null = null,
    rectWidth: '15px',
    rectHeight: '15px',
    columnWidth: '50px',
) => {
    const base = target
        .style('display', 'flex')
        .style('align-items', 'center')
        .style('min-height', '33px')
        .style('margin-left', '15px');

    const container = base.append('div')
        .style('width', '100%')
        .style('columns', columnWidth)
        .style('cursor', 'pointer');

    container.selectAll('div')
        .data(names)
        .join('div')
        .on('click', (a: any, b: string) => { onClick(b); })
        .html(function (d) {
            const el = d3.select(this);
            const base = el.append('div')
                .style('break-inside', 'avoid')
                .style('display', 'flex')
                .style('align-items', 'center')
                .style('padding-bottom', '3px')
                .style('line-height', '1.8')
                .style('font', d => d === nameToHighlight ? '12px sans-serif' : '11px sans-serif')
                .style('font-weight', d => d === nameToHighlight ? 'bolder' : 'normal');

            // each square
            base.append('div')
                .style('width', rectWidth)
                .style('height', rectHeight)
                .style('margin', '0 0.5em 0 0')
                .style('border', '#d8d8d8')
                .style('background', colorMap[d]);

            // each label
            base.append('div')
                .style('white-space', 'nowrap')
                .style('overflow', 'hidden')
                .style('text-overflow', 'ellipsis')
                .style('max-width', 'calc(100% - 15px - 0.5em)')
                .attr('title', transformName(d))
                .html(transformName(d));

            return el.html();
        });
}
