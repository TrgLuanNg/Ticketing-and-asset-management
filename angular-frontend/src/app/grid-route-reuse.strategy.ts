import { RouteReuseStrategy, ActivatedRouteSnapshot, DetachedRouteHandle } from '@angular/router';

export class GridRouteReuseStrategy implements RouteReuseStrategy {
  private storedHandles = new Map<string, DetachedRouteHandle>();

  shouldDetach(route: ActivatedRouteSnapshot): boolean {
    return route.routeConfig?.path === '';
  }

  store(route: ActivatedRouteSnapshot, handle: DetachedRouteHandle): void {
    const key = this.getKey(route);
    this.storedHandles.set(key, handle);
    if (handle) {
      const element = (handle as any).componentRef?.location?.nativeElement;
      if (element) {
        const scrollBody = element.querySelector('.ant-table-body');
        if (scrollBody) {
          (handle as any)._scrollTop = scrollBody.scrollTop;
          (handle as any)._scrollLeft = scrollBody.scrollLeft;
        }
      }
    }
  }

  shouldAttach(route: ActivatedRouteSnapshot): boolean {
    const key = this.getKey(route);
    return this.storedHandles.has(key);
  }

  retrieve(route: ActivatedRouteSnapshot): DetachedRouteHandle | null {
    const key = this.getKey(route);
    const handle = this.storedHandles.get(key) ?? null;
    if (handle) {
      const element = (handle as any).componentRef?.location?.nativeElement;
      const scrollTop = (handle as any)._scrollTop;
      const scrollLeft = (handle as any)._scrollLeft;
      if (element && (scrollTop !== undefined || scrollLeft !== undefined)) {
        setTimeout(() => {
          const scrollBody = element.querySelector('.ant-table-body');
          if (scrollBody) {
            if (scrollTop !== undefined) {
              scrollBody.scrollTop = scrollTop;
            }
            if (scrollLeft !== undefined) {
              scrollBody.scrollLeft = scrollLeft;
            }
          }
        }, 80);
      }
    }
    return handle;
  }

  shouldReuseRoute(future: ActivatedRouteSnapshot, curr: ActivatedRouteSnapshot): boolean {
    return future.routeConfig === curr.routeConfig;
  }

  clearHandle(route: ActivatedRouteSnapshot): void {
    const key = this.getKey(route);
    this.storedHandles.delete(key);
  }

  private getKey(route: ActivatedRouteSnapshot): string {
    return route.routeConfig?.path || '';
  }
}
