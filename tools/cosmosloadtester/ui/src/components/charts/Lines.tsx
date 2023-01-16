import React from 'react';
import * as d3 from 'd3';

import { scaleType } from 'src/types';

type Scaler = d3.ScaleContinuousNumeric<any, any, any>;
type groupEl = d3.Selection<SVGGElement, unknown, null, undefined>;
type divEl = d3.Selection<HTMLDivElement, unknown, null, undefined>;

export interface LinePoint {
    getX: () => number;
    getY: () => number;
    getDimension: () => string;
    getXAssociatedData: () => { [key: string]: number|string } | undefined;
};

interface LinesProp {
    target: React.RefObject<HTMLDivElement>;
    data: LinePoint[];
    dimensionToColorMap: { [key: string]: string };
    height: number;
    width: number;
    nameToHighlight: string|null;
    onHighlightName: (dimension: string) => void;
    xAxisTitle: string;
    yAxisTitle: string;
    xAxisShortTitle: string;
    yAxisShortTitle: string;
    yScale: any;
}

export default function Lines(props: LinesProp) {
    const {
        data,
        target,
        width,
        height,
        xAxisTitle,
        yAxisTitle,
        dimensionToColorMap,
        nameToHighlight,
        // onHighlightName,
        xAxisShortTitle,
        yAxisShortTitle,
        yScale,
    } = props;

    React.useEffect(() => {
        if (!data.length) return;
        d3.select(target.current).select('svg').remove();
        const svg = d3.select(target.current)
            .append('svg')
            .attr('width', width)
            .attr('height', height)
            .style('overflow', 'visible');
    
        const margin = { left: 40, right: 30, top: 30, bottom: 40 };
    
        const linesData: { [dim: string]: LinePoint[] } = {};
    
        let maxYAxisVal = data[0].getY();
        let minYAxisVal = maxYAxisVal;
    
        let minXAxisVal = data[0].getX();
        let maxXAxisVal = minXAxisVal;
        
        data.forEach((d) => {
            const dim = d.getDimension();
            if (linesData[dim] == null) linesData[dim] = [];
            const x = d.getX();
            const y = d.getY();

            if (x < minXAxisVal) minXAxisVal = x;
            if (x > maxXAxisVal) maxXAxisVal = x;
    
            if (y > maxYAxisVal) maxYAxisVal = y;
            if (y < minYAxisVal) minYAxisVal = y;
    
            linesData[dim].push(d);
        });
    
        const getYMax = (): number => {
            if (nameToHighlight) {
                const key = nameToHighlight;
                const points = linesData[key];
                const max = d3.max(points, d => d.getY()) as number;
                return max;
            }
            return maxYAxisVal;
        };
    
        const getYMin = (): number => {
            if (nameToHighlight) {
                const key = nameToHighlight;
                const points = linesData[key];
                const min = d3.min(points, d => d.getY()) as number;
                return min;
            }
            return minYAxisVal;
        }
    
        const x = d3.scaleLinear<number>()
            .domain([1, maxXAxisVal])
            .range([margin.left, width - margin.right])
            .nice()
            .clamp(true);
    
        const yLinear = d3.scaleLinear<number>()
            .domain([getYMin(), getYMax()])
            .range([height - margin.bottom, margin.top])
            .nice()
            .clamp(true);
    
        const yLog = d3.scaleLog<number>()
            .base(2)
            .domain([getYMin(), getYMax()])
            .rangeRound([height - margin.bottom, margin.top])
            .nice()
            .clamp(true);
    
        const xAxis = (g: groupEl) => g
            .attr('transform', `translate(0, ${height - margin.bottom})`)
            .call(d3.axisBottom(x).ticks(width / 60))
            .call(g => g.select('.domain').attr('opacity', 0.2))
            .call(g => g.append('text')
                .attr('fill', 'currentColor')
                .attr('text-anchor', 'start')
                .attr('opacity', 0.5)
                .style('font', '10px sans-serif')
                .attr('transform', `translate(${width / 2 - 100}, ${margin.bottom / 1.2})`)
                .text(xAxisTitle));
    
        const yAxis = (g: groupEl, ys: Scaler) => g
            .attr('transform', `translate(${margin.left}, 0)`)
            .call(d3.axisLeft(ys).ticks(height / 60, '~s'))
            .call(g => g.selectAll('.tick line').clone()
                .attr('stroke-dasharray', '1.5,2')
                .attr('stroke-opacity', 0.2)
                .attr('x2', width - margin.left - margin.right))
            .call(g => g.select('.domain').remove())
            .call(g => g.append('text')
                .attr('x', -40)
                .attr('y', margin.top - 20)
                .attr('fill', 'currentColor')
                .attr('text-anchor', 'start')
                .style('font', '10px sans-serif')
                .text(yAxisTitle));
    
        const line = (ys: Scaler) => d3.line<LinePoint>()
            .x(d => x(d.getX()))
            .y(d => ys(d.getY()));
    
        const paths = (g: groupEl, ys: Scaler) => {
            g.attr('fill', 'none')
                .attr('stroke-width', '1.5')
                .style('cursor', 'pointer')
                .attr('stroke-linejoin', 'round')
                .attr('stroke-linecap', 'round')
            .selectAll('path')
                .data(Object.entries(linesData))
                .join('path')
                .attr('stroke', ([k]) => dimensionToColorMap[k])
                .attr('d', ([, v]) => line(ys)(v))
                .style('opacity', ([k]) => {
                    if (nameToHighlight === null) return 1;
                    else if (k === nameToHighlight) return 1;
                    else return 0;
                });
        };
    
        const createTooltip = (el: divEl) => {
            el
                .style("position", "absolute")
                .style("pointer-events", "none")
                .style("top", 0)
                .style("opacity", 0)
                .style("background", "white")
                .style("border-radius", "5px")
                .style("box-shadow", "0 0 10px rgba(0,0,0,.25)")
                .style("padding", "10px")
                .style("line-height", "1.3")
                .style("font", "11px sans-serif");
        }
    
        const tooltipContent = (key: string, d: LinePoint) => {
            const color = d3.color(dimensionToColorMap[d.getDimension()])?.darker() ?? '#555';
            let str = '';
            const assoc = d.getXAssociatedData?.();
            if (assoc) {
                Object.keys(assoc).forEach((k) => {
                    str += `<span style="color:${color}">${k}: ${assoc[k]}</span> <br />`
                });
            } else {
                str += `<b style="color:${color}">${xAxisShortTitle}: ${d.getX()}</b><br/>`;
            }
            return (`
                ${str}<b style="font-weight:400">${yAxisShortTitle}: ${Math.round(d.getY()).toLocaleString()}</b> <br />
            `);
        }
    
        d3.select(target.current).select('div.tooltip').remove();
        const tooltip = d3.select(target.current).append('div').attr('class', 'tooltip').call(createTooltip);
    
        const dots = (g: groupEl, ys: Scaler) => {
            g.selectAll('g')
                .data(Object.entries(linesData))
                .join('g')
                .attr('fill', ([k]) => dimensionToColorMap[k])
                .each(function ([key, points]) {
                    const group = d3.select(this).append('g')
                    group.selectAll('circle')
                        .data(points)
                        .join('circle')
                        .style('cursor', 'pointer')
                        .attr('r', () => {
                            if (nameToHighlight === null) return 2.5;
                            else if (nameToHighlight === key) return 3;
                            else return 0;
                        })
                        .attr('cx', (d) => x(d.getX()))
                        .attr('cy', (d) => ys(d.getY()))
                        .on("touchmove mousemove", (_, d) => {
                            if (nameToHighlight === null || key === nameToHighlight) {
                                tooltip.style("opacity", 1).html(tooltipContent(key, d));
                            }
                        })
                        .on("touchend mouseleave", () => {
                            if (nameToHighlight === null || key === nameToHighlight) {
                                tooltip.style("opacity", 0);
                            }
                        });
                });
        };
    
        svg.on('mousemove', function (event) {
            let [x, y] = d3.pointer(event, svg);
            tooltip.style('left', `${x}px`).style('top', `${y}px`);
        });
    
        const yScaleMap: Map<scaleType, Scaler> = new Map([
            ['linear', yLinear],
            ['log', yLog],
        ]);
    
        const y = yScaleMap.get(yScale) as Scaler;
    
        svg.append('g').call(yAxis, y);
        svg.append('g').call(xAxis);
        svg.append('g').call(paths, y);
        svg.append('g').call(dots, y);
    }, [
        target, width, height, yScale, nameToHighlight, dimensionToColorMap,
        data, yAxisShortTitle, yAxisTitle, xAxisShortTitle, xAxisTitle,
    ]);

    return null;
}
