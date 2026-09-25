import { Injectable, TemplateRef } from '@angular/core';
import { NzMessageDataOptions, NzMessageRef } from 'ng-zorro-antd/message';

@Injectable({
  providedIn: 'root'
})
export class CustomMessageService {
  private container: HTMLElement | null = null;

  constructor() {
    this.createContainer();
  }

  private createContainer() {
    if (typeof document === 'undefined') return;
    this.container = document.getElementById('native-toast-container');
    if (!this.container) {
      this.container = document.createElement('div');
      this.container.id = 'native-toast-container';
      this.container.style.position = 'fixed';
      this.container.style.top = '24px';
      this.container.style.left = '50%';
      this.container.style.transform = 'translateX(-50%)';
      this.container.style.zIndex = '999999';
      this.container.style.display = 'flex';
      this.container.style.flexDirection = 'column';
      this.container.style.alignItems = 'center';
      this.container.style.gap = '10px';
      document.body.appendChild(this.container);
    }
  }

  private showToast(message: string, type: 'success' | 'error' | 'warning' | 'info', duration: number = 3000) {
    if (typeof document === 'undefined') return;
    if (!this.container) this.createContainer();
    if (!this.container) return;

    const toast = document.createElement('div');
    toast.style.padding = '14px 28px';
    toast.style.borderRadius = '8px';
    toast.style.color = '#ffffff';
    toast.style.fontSize = '14px';
    toast.style.fontWeight = '600';
    toast.style.boxShadow = '0 10px 15px -3px rgba(0, 0, 0, 0.15), 0 4px 6px -2px rgba(0, 0, 0, 0.05)';
    toast.style.display = 'flex';
    toast.style.alignItems = 'center';
    toast.style.gap = '10px';
    toast.style.minWidth = '320px';
    toast.style.maxWidth = '500px';
    toast.style.fontFamily = "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif";
    toast.style.transition = 'all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275)';
    toast.style.transform = 'translateY(-50px)';
    toast.style.opacity = '0';

    // Type styling
    if (type === 'success') {
      toast.style.backgroundColor = '#10b981';
      toast.style.borderLeft = '4px solid #047857';
    } else if (type === 'error') {
      toast.style.backgroundColor = '#ef4444';
      toast.style.borderLeft = '4px solid #b91c1c';
    } else if (type === 'warning') {
      toast.style.backgroundColor = '#f59e0b';
      toast.style.borderLeft = '4px solid #b45309';
    } else {
      toast.style.backgroundColor = '#3b82f6';
      toast.style.borderLeft = '4px solid #1d4ed8';
    }

    toast.innerText = message;
    this.container.appendChild(toast);

    // Fade and slide in
    setTimeout(() => {
      toast.style.transform = 'translateY(0)';
      toast.style.opacity = '1';
    }, 50);

    // Slide out and remove
    setTimeout(() => {
      toast.style.transform = 'translateY(-50px)';
      toast.style.opacity = '0';
      setTimeout(() => {
        toast.remove();
      }, 350);
    }, duration);
  }

  success(content: string | TemplateRef<void>, options?: NzMessageDataOptions): NzMessageRef {
    const text = typeof content === 'string' ? content : 'Thành công';
    this.showToast(text, 'success', options?.nzDuration || 3000);
    return {} as NzMessageRef;
  }

  error(content: string | TemplateRef<void>, options?: NzMessageDataOptions): NzMessageRef {
    const text = typeof content === 'string' ? content : 'Lỗi xảy ra';
    this.showToast(text, 'error', options?.nzDuration || 4000);
    return {} as NzMessageRef;
  }

  warning(content: string | TemplateRef<void>, options?: NzMessageDataOptions): NzMessageRef {
    const text = typeof content === 'string' ? content : 'Cảnh báo';
    this.showToast(text, 'warning', options?.nzDuration || 4000);
    return {} as NzMessageRef;
  }

  info(content: string | TemplateRef<void>, options?: NzMessageDataOptions): NzMessageRef {
    const text = typeof content === 'string' ? content : 'Thông tin';
    this.showToast(text, 'info', options?.nzDuration || 3000);
    return {} as NzMessageRef;
  }

  loading(content: string | TemplateRef<void>, options?: NzMessageDataOptions): NzMessageRef {
    const text = typeof content === 'string' ? content : 'Đang tải...';
    this.showToast(text, 'info', options?.nzDuration || 3000);
    return {} as NzMessageRef;
  }

  remove(messageId?: string): void {
    if (this.container) {
      this.container.innerHTML = '';
    }
  }

  create(type: 'success' | 'info' | 'warning' | 'error' | 'loading', content: string | TemplateRef<void>, options?: NzMessageDataOptions): NzMessageRef {
    switch (type) {
      case 'success': return this.success(content, options);
      case 'error': return this.error(content, options);
      case 'warning': return this.warning(content, options);
      case 'info': return this.info(content, options);
      case 'loading': return this.loading(content, options);
      default: return {} as NzMessageRef;
    }
  }
}
