import { Component, OnInit, ViewChild, ElementRef, ChangeDetectorRef, AfterViewInit, OnDestroy, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzCardModule } from 'ng-zorro-antd/card';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { AssetService } from '../../core/services/asset.service';
import * as echarts from 'echarts';

import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzModalModule } from 'ng-zorro-antd/modal';

@Component({
  selector: 'app-statistics',
  standalone: true,
  imports: [
    CommonModule, 
    NzGridModule, 
    NzCardModule, 
    NzSelectModule, 
    NzSpinModule,
    NzIconModule,
    NzButtonModule,
    NzModalModule,
    FormsModule
  ],
  templateUrl: './statistics.html',
  styleUrl: './statistics.less',
})
export class Statistics implements OnInit, AfterViewInit, OnDestroy {
  @ViewChild('groupChart', { static: false }) groupChartEl!: ElementRef;
  @ViewChild('regionChart', { static: false }) regionChartEl!: ElementRef;
  @ViewChild('divisionChart', { static: false }) divisionChartEl!: ElementRef;
  @ViewChild('stateChart', { static: false }) stateChartEl!: ElementRef;
  @ViewChild('eolChart', { static: false }) eolChartEl!: ElementRef;
  @ViewChild('purchasedChart', { static: false }) purchasedChartEl!: ElementRef;
  @ViewChild('modalBarChart', { static: false }) modalBarChartEl?: ElementRef;

  private groupChartInstance: echarts.ECharts | null = null;
  private regionChartInstance: echarts.ECharts | null = null;
  private divisionChartInstance: echarts.ECharts | null = null;
  private stateChartInstance: echarts.ECharts | null = null;
  private eolChartInstance: echarts.ECharts | null = null;
  private purchasedChartInstance: echarts.ECharts | null = null;
  private modalChartInstance: echarts.ECharts | null = null;

  public loading = false;
  public totalAssets = 0;
  public transferPC = 0;
  public newPC = 0;

  public groupData: any[] = [];
  public regionData: any[] = [];
  public divisionData: any[] = [];
  public stateData: any[] = [];

  public isModalVisible = false;
  public modalTitle = '';
  public modalWidth = '700px';
  private currentModalData: any[] = [];

  monthOptions: { label: string, value: string }[] = [];
  selectedFromDate = '';

  constructor(
    private assetService: AssetService,
    private cdr: ChangeDetectorRef,
    private router: Router
  ) {}

  ngOnInit() {
    this.generateMonthOptions();
    this.loadDashboardData();
  }

  ngAfterViewInit() {
    // Initial resize listener
    window.addEventListener('resize', this.resizeCharts);
  }

  ngOnDestroy() {
    window.removeEventListener('resize', this.resizeCharts);
    this.disposeCharts();
  }

  generateMonthOptions() {
    const options = [];
    const startDate = new Date(2018, 0, 1);
    const endDate = new Date();
    
    let current = new Date(endDate.getFullYear(), endDate.getMonth(), 1);
    while (current >= startDate) {
      const month = String(current.getMonth() + 1).padStart(2, '0');
      const year = current.getFullYear();
      const val = `${month}/${year}`;
      options.push({ label: val, value: val });
      current.setMonth(current.getMonth() - 1);
    }
    this.monthOptions = options;
    this.selectedFromDate = '04/2026';
  }

  onFilterChange() {
    this.loadDashboardData();
  }

  loadDashboardData() {
    this.loading = true;
    this.assetService.getDashboardMetrics(this.selectedFromDate).subscribe({
      next: (res: any) => {
        this.loading = false;
        if (res.status && res.data) {
          const d = res.data;
          this.totalAssets = d.totalAssets;
          this.transferPC = d.transferPC;
          this.newPC = d.newPC;

          this.groupData = d.groupData || [];
          this.regionData = d.regionData || [];
          this.divisionData = d.divisionData || [];
          this.stateData = d.stateData || [];

          // Render all ECharts after DOM has updated
          setTimeout(() => {
            this.renderGroupChart(this.groupData);
            this.renderRegionChart(this.regionData);
            this.renderDivisionChart(this.divisionData);
            this.renderStateChart(this.stateData);
            this.renderEolChart(d.eolData);
            this.renderPurchasedChart(d.purchasedData);
            this.cdr.detectChanges();
          }, 50);
        }
        this.cdr.detectChanges();
      },
      error: (err) => {
        console.error('Failed to load dashboard metrics', err);
        this.loading = false;
        this.cdr.detectChanges();
      }
    });
  }

  navigateToAssets(target?: string) {
    if (target === 'new' && this.selectedFromDate) {
      this.router.navigate(['/assets'], { queryParams: { purchasedFrom: this.selectedFromDate } });
    } else {
      this.router.navigate(['/assets']);
    }
  }

  private resizeCharts = () => {
    this.groupChartInstance?.resize();
    this.regionChartInstance?.resize();
    this.divisionChartInstance?.resize();
    this.stateChartInstance?.resize();
    this.eolChartInstance?.resize();
    this.purchasedChartInstance?.resize();
  };

  private disposeCharts() {
    if (this.groupChartInstance) { this.groupChartInstance.dispose(); this.groupChartInstance = null; }
    if (this.regionChartInstance) { this.regionChartInstance.dispose(); this.regionChartInstance = null; }
    if (this.divisionChartInstance) { this.divisionChartInstance.dispose(); this.divisionChartInstance = null; }
    if (this.stateChartInstance) { this.stateChartInstance.dispose(); this.stateChartInstance = null; }
    if (this.eolChartInstance) { this.eolChartInstance.dispose(); this.eolChartInstance = null; }
    if (this.purchasedChartInstance) { this.purchasedChartInstance.dispose(); this.purchasedChartInstance = null; }
  }

  private renderGroupChart(data: any[]) {
    if (!this.groupChartEl) return;
    if (!this.groupChartInstance) {
      this.groupChartInstance = echarts.init(this.groupChartEl.nativeElement);
    }
    const option = this.getDonutOption('Group Distribution', data);
    this.groupChartInstance.setOption(option);
  }

  private renderRegionChart(data: any[]) {
    if (!this.regionChartEl) return;
    if (!this.regionChartInstance) {
      this.regionChartInstance = echarts.init(this.regionChartEl.nativeElement);
    }
    const option = this.getDonutOption('Region Distribution', data);
    this.regionChartInstance.setOption(option);
  }

  private renderDivisionChart(data: any[]) {
    if (!this.divisionChartEl) return;
    if (!this.divisionChartInstance) {
      this.divisionChartInstance = echarts.init(this.divisionChartEl.nativeElement);
    }
    const option = this.getDonutOption('Division Distribution', data);
    this.divisionChartInstance.setOption(option);
  }

  private renderStateChart(data: any[]) {
    if (!this.stateChartEl) return;
    if (!this.stateChartInstance) {
      this.stateChartInstance = echarts.init(this.stateChartEl.nativeElement);
    }
    const option = this.getDonutOption('PC State', data);
    this.stateChartInstance.setOption(option);
  }

  private renderEolChart(data: any[]) {
    if (!this.eolChartEl) return;
    if (!this.eolChartInstance) {
      this.eolChartInstance = echarts.init(this.eolChartEl.nativeElement);
    }
    // Transform to ECharts Treemap format
    const treemapData = data.map(x => ({
      name: x.name,
      value: x.value
    }));

    const option = {
      tooltip: {
        trigger: 'item',
        formatter: '{b}: {c} PCs'
      },
      series: [
        {
          name: 'EOL PC',
          type: 'treemap',
          silent: true,
          roam: false,
          nodeClick: false,
          visibleMinArea: 300,
          label: {
            show: true,
            formatter: '{b}\n\n{c}',
            fontSize: 14,
            fontWeight: 'bold',
            color: '#fff'
          },
          itemStyle: {
            borderColor: '#fff',
            borderWidth: 2
          },
          breadcrumb: {
            show: false
          },
          levels: [
            {
              itemStyle: {
                borderWidth: 2,
                gapWidth: 2
              }
            }
          ],
          data: treemapData
        }
      ]
    };
    this.eolChartInstance.setOption(option);
  }

  private renderPurchasedChart(data: any[]) {
    if (!this.purchasedChartEl) return;
    if (!this.purchasedChartInstance) {
      this.purchasedChartInstance = echarts.init(this.purchasedChartEl.nativeElement);
    }
    const years = data.map(x => x.year);
    const counts = data.map(x => x.count);

    // Color palette to match draftStats.png
    const customColors = ['#82a563', '#d9d2b6', '#d6be5e', '#bca333', '#e0ecf8', '#c9a12c', '#1890ff', '#e6f7ff'];

    const option = {
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'shadow' }
      },
      grid: {
        left: '2%',
        right: '2%',
        bottom: '2%',
        top: '12%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        data: years,
        axisLine: { lineStyle: { color: '#cbd5e1' } },
        axisLabel: { color: '#64748b', fontWeight: 'bold' }
      },
      yAxis: {
        type: 'value',
        axisLine: { show: false },
        splitLine: { lineStyle: { type: 'dashed', color: '#f1f5f9' } },
        axisLabel: { color: '#64748b' }
      },
      series: [
        {
          name: 'Purchased Count',
          type: 'bar',
          barWidth: '50%',
          data: counts,
          label: {
            show: true,
            position: 'top',
            color: '#475569',
            fontWeight: 'bold'
          },
          itemStyle: {
            borderRadius: [4, 4, 0, 0],
            color: (params: any) => {
              const yearVal = years[params.dataIndex]?.toString();
              if (yearVal === '2022') {
                return '#3D2B06';
              }
              if (yearVal === '2025') {
                return '#063D0F';
              }
              return customColors[params.dataIndex % customColors.length];
            }
          }
        }
      ]
    };
    this.purchasedChartInstance.setOption(option);
  }

  private getDonutOption(title: string, data: any[]) {
    return {
      tooltip: {
        trigger: 'item',
        formatter: '{b}: {c} ({d}%)'
      },
      legend: {
        show: false
      },
      series: [
        {
          name: title,
          type: 'pie',
          radius: ['65%', '75%'],
          avoidLabelOverlap: true,
          label: {
            show: true,
            position: 'outside',
            formatter: '{b}\n{c}',
            color: '#475569',
            fontSize: 11
          },
          labelLine: {
            show: true,
            length: 8,
            length2: 8,
            lineStyle: {
              color: '#cbd5e1'
            }
          },
          data: data
        }
      ]
    };
  }

  openBarChartModal(title: string, data: any[]) {
    this.modalTitle = `${title} Bar Chart`;
    this.currentModalData = data;
    this.modalWidth = Math.max(700, (data ? data.length : 0) * 60) + 'px';
    this.isModalVisible = true;
    
    setTimeout(() => {
      if (this.modalBarChartEl) {
        if (this.modalChartInstance) {
          this.modalChartInstance.dispose();
          this.modalChartInstance = null;
        }
        this.modalChartInstance = echarts.init(this.modalBarChartEl.nativeElement);
        
        const categories = this.currentModalData.map(x => x.name);
        const values = this.currentModalData.map(x => x.value);
        
        const option = {
          tooltip: {
            trigger: 'axis',
            axisPointer: { type: 'shadow' }
          },
          grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            top: '10%',
            containLabel: true
          },
          xAxis: {
            type: 'category',
            data: categories,
            axisLabel: {
              interval: 0,
              rotate: 30,
              color: '#64748b'
            }
          },
          yAxis: {
            type: 'value',
            axisLabel: { color: '#64748b' }
          },
          series: [
            {
              name: title,
              type: 'bar',
              barWidth: '40%',
              data: values,
              label: {
                show: true,
                position: 'top',
                color: '#475569',
                fontWeight: 'bold'
              },
              itemStyle: {
                borderRadius: [4, 4, 0, 0],
                color: '#1890ff'
              }
            }
          ]
        };
        this.modalChartInstance.setOption(option);
      }
    }, 200);
  }

  handleModalClose() {
    this.isModalVisible = false;
    if (this.modalChartInstance) {
      this.modalChartInstance.dispose();
      this.modalChartInstance = null;
    }
  }

  @HostListener('window:resize')
  onResize() {
    this.groupChartInstance?.resize();
    this.regionChartInstance?.resize();
    this.divisionChartInstance?.resize();
    this.stateChartInstance?.resize();
    this.eolChartInstance?.resize();
    this.purchasedChartInstance?.resize();
    this.modalChartInstance?.resize();
  }
}
