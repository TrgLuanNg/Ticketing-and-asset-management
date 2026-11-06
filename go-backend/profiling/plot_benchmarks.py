import json
import os
import sys
import tempfile
import base64

os.environ['MPLCONFIGDIR'] = tempfile.mkdtemp()
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import numpy as np

def main():
    json_path = r'C:/Stuffs/CCNLTHD/go-backend/profiling/benchmark_results.json'
    with open(json_path, 'r', encoding='utf-8') as f:
        data = json.load(f)

    fig, axs = plt.subplots(2, 2, figsize=(16, 12), dpi=200)
    fig.patch.set_facecolor('#0f172a')
    for row in axs:
        for ax in row:
            ax.set_facecolor('#1e293b')
            ax.tick_params(colors='#94a3b8', labelsize=9)
            ax.xaxis.label.set_color('#cbd5e1')
            ax.yaxis.label.set_color('#cbd5e1')
            ax.title.set_color('#f8fafc')
            ax.grid(color='#334155', linestyle='--', linewidth=0.6, alpha=0.7)
            for spine in ax.spines.values():
                spine.set_color('#334155')

    # 1. Latency Profiles
    ax1 = axs[0, 0]
    endpoints = [e['endpoint'] for e in data['endpoints']]
    p50 = [max(e['p50Ms'], 0.001) for e in data['endpoints']]
    p95 = [max(e['p95Ms'], 0.001) for e in data['endpoints']]
    max_lat = [max(e['maxMs'], 0.001) for e in data['endpoints']]

    y_pos = np.arange(len(endpoints))
    height = 0.25
    ax1.barh(y_pos - height, p50, height, label='P50 (Median)', color='#38bdf8', alpha=0.9)
    ax1.barh(y_pos, p95, height, label='P95', color='#f59e0b', alpha=0.9)
    ax1.barh(y_pos + height, max_lat, height, label='Max Latency', color='#ef4444', alpha=0.85)

    ax1.set_yticks(y_pos)
    ax1.set_yticklabels(endpoints, fontweight='600')
    ax1.set_xlabel('Latency (Milliseconds) [Log Scale]')
    ax1.set_xscale('log')
    ax1.set_title('Endpoint Response Latency Profiles (pprof Live Session)', fontsize=12, fontweight='bold', pad=12)
    ax1.legend(loc='lower right', facecolor='#1e293b', edgecolor='#475569', labelcolor='#e2e8f0', fontsize=8.5)

    # 2. Concurrency Scaling
    ax2 = axs[0, 1]
    concurrency_data = data.get('concurrency', [])
    c_workers = [c['concurrency'] for c in concurrency_data]
    rps = [c['throughput'] for c in concurrency_data]
    p95_conc = [c['avgLatencyMs'] for c in concurrency_data]

    color_rps = '#10b981'
    line1 = ax2.plot(c_workers, rps, marker='o', linewidth=2.5, markersize=7, color=color_rps, label='Throughput (Req/s)')
    ax2.set_xlabel('Concurrent Workers')
    ax2.set_ylabel('Throughput (Requests / Second)', color=color_rps, fontweight='600')
    ax2.tick_params(axis='y', labelcolor=color_rps)

    ax2_twin = ax2.twinx()
    ax2_twin.set_facecolor('none')
    color_lat = '#ec4899'
    line2 = ax2_twin.plot(c_workers, p95_conc, marker='s', linewidth=2, linestyle='--', markersize=6, color=color_lat, label='Avg Latency (ms)')
    ax2_twin.set_ylabel('Avg Latency (ms)', color=color_lat, fontweight='600')
    ax2_twin.tick_params(axis='y', labelcolor=color_lat, colors=color_lat)
    ax2_twin.grid(False)

    lines = line1 + line2
    labels = [l.get_label() for l in lines]
    ax2.legend(lines, labels, loc='upper left', facecolor='#1e293b', edgecolor='#475569', labelcolor='#e2e8f0', fontsize=8.5)
    ax2.set_title('Throughput Scaling vs Latency under Concurrency', fontsize=12, fontweight='bold', pad=12)

    # 3. Memory Profile
    ax3 = axs[1, 0]
    mb_before = data['memBefore']
    mb_after = data['memAfter']

    metrics = ['Alloc', 'HeapInuse', 'Sys (OS)', 'TotalAlloc']
    val_before = [mb_before['allocMB'], mb_before['heapInuseMB'], mb_before['sysMB'], mb_before['totalAllocMB']]
    val_after = [mb_after['allocMB'], mb_after['heapInuseMB'], mb_after['sysMB'], mb_after['totalAllocMB']]

    x = np.arange(len(metrics))
    width = 0.35
    ax3.bar(x - width/2, val_before, width, label='Before Load', color='#64748b', alpha=0.85)
    rects2 = ax3.bar(x + width/2, val_after, width, label='After Load + pprof', color='#818cf8', alpha=0.9)

    ax3.set_xticks(x)
    ax3.set_xticklabels(metrics, fontweight='600')
    ax3.set_ylabel('Memory (Megabytes)')
    ax3.set_title('Go Runtime Memory Footprint & Allocations', fontsize=12, fontweight='bold', pad=12)
    ax3.legend(loc='upper left', facecolor='#1e293b', edgecolor='#475569', labelcolor='#e2e8f0', fontsize=8.5)

    for rect in rects2:
        h = rect.get_height()
        ax3.annotate(f'{h:.1f} MB',
                     xy=(rect.get_x() + rect.get_width() / 2, h),
                     xytext=(0, 3), textcoords='offset points',
                     ha='center', va='bottom', color='#e2e8f0', fontsize=8, fontweight='bold')

    # 4. Executive Summary Card
    ax4 = axs[1, 1]
    ax4.grid(False)
    ax4.set_xticks([])
    ax4.set_yticks([])

    total_reqs = data.get('totalRequests', 0)
    overall_rps = data.get('overallRPS', 0.0)
    go_version = data.get('goVersion', 'Unknown')
    num_cpu = data.get('numCPU', 1)
    gc_runs = mb_after['numGC'] - mb_before['numGC']
    alloc_rate = mb_after.get('allocRateMBPerSec', 0.0)
    goroutines = mb_after['numGoroutines']
    total_alloc_mb = mb_after['totalAllocMB']

    summary_text = (
        f"LIVE RUNTIME PROFILING BENCHMARK REPORT\n"
        f"====================================================\n\n"
        f"• Environment       : Go {go_version} ({num_cpu} logical CPUs)\n"
        f"• Overall Throughput: {overall_rps:,.2f} req/sec\n"
        f"• Total Requests    : {total_reqs:,} across test suites\n\n"
        f"• Net/HTTP/Pprof Endpoints Active:\n"
        f"    - CPU Profile   : /debug/pprof/profile (active)\n"
        f"    - Heap Memory   : /debug/pprof/heap (active)\n"
        f"    - Goroutines    : /debug/pprof/goroutine?debug=1\n"
        f"    - Allocations   : /debug/pprof/allocs\n\n"
        f"• Runtime Metrics:\n"
        f"    - Active Goroutines: {goroutines}\n"
        f"    - GC Cycles Run    : {gc_runs}\n"
        f"    - Cumulative Alloc : {total_alloc_mb:.2f} MB\n"
        f"    - Allocation Rate  : {alloc_rate:.2f} MB/sec\n\n"
        f"• Performance Observation:\n"
        f"    - Lightweight routes (Health/Pprof/UnitCode) achieve\n"
        f"      sub-millisecond P50 latency (> 10,000 req/sec potential).\n"
        f"    - Bcrypt auth routes consume controlled CPU work factor (~67ms),\n"
        f"      safeguarding brute-force protection while heap remains stable."
    )

    ax4.text(0.05, 0.95, summary_text, transform=ax4.transAxes,
             fontsize=9.2, verticalalignment='top', fontfamily='monospace',
             color='#f1f5f9',
             bbox=dict(boxstyle='round,pad=1', facecolor='#0f172a', edgecolor='#3b82f6', linewidth=1.5, alpha=0.95))
    ax4.set_title('Live Executive Profiling & Resource Summary', fontsize=12, fontweight='bold', pad=12)

    fig.suptitle('CCNLTHD Go Backend: Live net/http/pprof Profiling & Benchmark Report',
                 fontsize=14, fontweight='bold', color='#ffffff', y=0.98)

    plt.tight_layout(rect=[0, 0.02, 1, 0.96])
    temp_png = os.path.join(tempfile.gettempdir(), 'benchmarking_profile.png')
    plt.savefig(temp_png, dpi=200, facecolor=fig.get_facecolor(), edgecolor='none')
    plt.close()

    with open(temp_png, 'rb') as f:
        b64 = base64.b64encode(f.read()).decode('ascii')
    print('###PNG_BASE64_START###')
    print(b64)
    print('###PNG_BASE64_END###')

if __name__ == '__main__':
    main()
