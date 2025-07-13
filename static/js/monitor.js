// 监控面板 JavaScript
class LLMMonitor {
    constructor() {
        this.providers = [];
        this.refreshInterval = 30000; // 30秒
        this.refreshTimer = null;
        this.init();
    }

    init() {
        this.bindEvents();
        this.loadData();
        this.startAutoRefresh();
    }

    bindEvents() {
        // 刷新按钮
        document.getElementById('refresh-btn').addEventListener('click', () => {
            this.loadData();
        });

        // 测试按钮
        document.getElementById('test-btn').addEventListener('click', () => {
            this.testProvider();
        });

        // 提供商选择变化
        document.getElementById('test-provider').addEventListener('change', (e) => {
            this.updateModelOptions(e.target.value);
        });

        // 模态框关闭
        document.querySelector('.close').addEventListener('click', () => {
            this.closeModal();
        });

        // 点击模态框外部关闭
        window.addEventListener('click', (e) => {
            const modal = document.getElementById('modal');
            if (e.target === modal) {
                this.closeModal();
            }
        });
    }

    async loadData() {
        this.showLoading();
        try {
            await Promise.all([
                this.loadProviders(),
                this.loadSystemStats()
            ]);
            this.updateLastRefreshTime();
        } catch (error) {
            console.error('加载数据失败:', error);
            this.showError('数据加载失败: ' + error.message);
        } finally {
            this.hideLoading();
        }
    }

    async loadProviders() {
        try {
            const response = await fetch('/admin/api/providers');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            
            if (data.success) {
                this.providers = data.providers;
                this.updateProvidersDisplay();
                this.updateProviderSelect();
                this.updateProviderStats();
            } else {
                throw new Error(data.error || '获取提供商数据失败');
            }
        } catch (error) {
            console.error('加载提供商数据失败:', error);
            throw error;
        }
    }

    async loadSystemStats() {
        try {
            const response = await fetch('/admin/api/stats');
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            const data = await response.json();
            
            if (data.success) {
                this.updateSystemStats(data.data);
            } else {
                throw new Error(data.error || '获取系统统计失败');
            }
        } catch (error) {
            console.error('加载系统统计失败:', error);
            throw error;
        }
    }

    updateProvidersDisplay() {
        const grid = document.getElementById('providers-grid');
        grid.innerHTML = '';

        this.providers.forEach(provider => {
            const card = this.createProviderCard(provider);
            grid.appendChild(card);
        });
    }

    createProviderCard(provider) {
        const card = document.createElement('div');
        card.className = `provider-card ${provider.status}`;
        
        const modelsCount = Array.isArray(provider.models) ? provider.models.length : 0;
        const lastTest = provider.lastTest ? new Date(provider.lastTest).toLocaleString() : '未测试';
        
        card.innerHTML = `
            <div class="provider-header">
                <div class="provider-name">${provider.name.toUpperCase()}</div>
                <span class="provider-status status-${provider.status}">${this.getStatusText(provider.status)}</span>
            </div>
            <div class="provider-summary">
                <div class="summary-item">
                    <i class="fas fa-layer-group"></i>
                    <span>${modelsCount} 个模型</span>
                </div>
                <div class="summary-item">
                    <i class="fas fa-chart-bar"></i>
                    <span>请求次数: ${provider.requests || 0}</span>
                </div>
                <div class="summary-item">
                    <i class="fas fa-coins"></i>
                    <span>Token消耗: ${provider.tokens || 0}</span>
                </div>
                <div class="summary-item">
                    <i class="fas fa-tachometer-alt"></i>
                    <span>平均响应: ${provider.avgResponseTime || 0}ms</span>
                </div>
            </div>
            <div class="provider-actions">
                <button class="btn btn-sm btn-success" onclick="monitor.testSpecificProvider('${provider.name}')">
                    <i class="fas fa-play"></i> 测试连接
                </button>
                <button class="btn btn-sm btn-warning" onclick="monitor.showCostStats('${provider.name}')">
                    <i class="fas fa-calculator"></i> 成本统计
                </button>
                <button class="btn btn-sm btn-primary" onclick="monitor.showProviderDetails('${provider.name}')">
                    <i class="fas fa-info-circle"></i> 查看详情
                </button>
            </div>
        `;

        return card;
    }

    updateProviderSelect() {
        const select = document.getElementById('test-provider');
        
        // 清除现有选项（保留"自动选择"）
        while (select.children.length > 1) {
            select.removeChild(select.lastChild);
        }

        // 添加提供商选项
        this.providers.forEach(provider => {
            const option = document.createElement('option');
            option.value = provider.name;
            option.textContent = provider.name.toUpperCase();
            select.appendChild(option);
        });
    }

    updateProviderStats() {
        const totalProviders = this.providers.length;
        const healthyProviders = this.providers.filter(p => p.status === 'healthy').length;

        document.getElementById('total-providers').textContent = totalProviders;
        document.getElementById('healthy-providers').textContent = healthyProviders;
    }

    updateSystemStats(stats) {
        // 更新仪表板统计
        document.getElementById('total-requests').textContent = stats.metrics.total_requests || 0;
        document.getElementById('avg-response-time').textContent = 
            (stats.metrics.avg_response_time || 0) + 'ms';

        // 更新服务信息
        document.getElementById('service-version').textContent = stats.service.version || 'v1.0.0';
        document.getElementById('uptime').textContent = stats.service.uptime || '--';
        document.getElementById('service-port').textContent = stats.service.port || 8080;

        // 更新实时指标
        document.getElementById('active-connections').textContent = 
            stats.metrics.active_connections || '--';
        document.getElementById('memory-usage').textContent = 
            stats.metrics.memory_usage || '--';
        document.getElementById('cpu-usage').textContent = 
            stats.metrics.cpu_usage || '--';
    }

    updateModelOptions(providerName) {
        const modelInput = document.getElementById('test-model');
        
        if (!providerName) {
            modelInput.placeholder = 'gpt-3.5-turbo';
            modelInput.value = '';
            return;
        }

        // 根据提供商设置默认模型
        const defaultModels = {
            'openai': 'gpt-3.5-turbo',
            'gemini': 'gemini-pro',
            'deepseek': 'deepseek-chat',
            'qwen': 'qwen-turbo',
            'moonshot': 'moonshot-v1-8k'
        };

        const defaultModel = defaultModels[providerName] || '';
        modelInput.placeholder = defaultModel;
        modelInput.value = defaultModel;
    }

    async testProvider() {
        const provider = document.getElementById('test-provider').value;
        const model = document.getElementById('test-model').value;
        const message = document.getElementById('test-message').value;

        if (!message.trim()) {
            this.showError('请输入测试消息');
            return;
        }

        const testBtn = document.getElementById('test-btn');
        const originalText = testBtn.innerHTML;
        testBtn.disabled = true;
        testBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 测试中...';

        const output = document.getElementById('test-output');
        output.textContent = '正在发送测试请求...';

        try {
            const response = await fetch('/admin/api/test', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    provider: provider,
                    model: model,
                    message: message
                })
            });

            const result = await response.json();
            
            if (result.success) {
                output.innerHTML = `<div style="color: #48bb78; margin-bottom: 10px;">✅ 测试成功</div>
<strong>提供商:</strong> ${result.provider}
<strong>模型:</strong> ${result.model}
<strong>响应时间:</strong> ${result.duration}ms
<strong>响应内容:</strong>
${JSON.stringify(result.response, null, 2)}`;
            } else {
                output.innerHTML = `<div style="color: #f56565; margin-bottom: 10px;">❌ 测试失败</div>
<strong>错误:</strong> ${result.error}
<strong>提供商:</strong> ${result.provider || 'N/A'}
<strong>响应时间:</strong> ${result.duration || 0}ms`;
            }
        } catch (error) {
            output.innerHTML = `<div style="color: #f56565; margin-bottom: 10px;">❌ 网络错误</div>
<strong>错误:</strong> ${error.message}`;
        } finally {
            testBtn.disabled = false;
            testBtn.innerHTML = originalText;
        }
    }

    async testSpecificProvider(providerName) {
        // 设置测试表单并执行测试
        document.getElementById('test-provider').value = providerName;
        this.updateModelOptions(providerName);
        await this.testProvider();
        
        // 滚动到测试结果
        document.querySelector('.test-section').scrollIntoView({ behavior: 'smooth' });
    }

    showProviderDetails(providerName) {
        const provider = this.providers.find(p => p.name === providerName);
        if (!provider) return;

        const modalTitle = document.getElementById('modal-title');
        const modalBody = document.getElementById('modal-body');

        modalTitle.textContent = `${provider.name.toUpperCase()} 详细信息`;
        
        const models = Array.isArray(provider.models) ? provider.models : [];
        
        modalBody.innerHTML = `
            <div style="margin-bottom: 20px;">
                <h4>基本信息</h4>
                <table style="width: 100%; border-collapse: collapse;">
                    <tr><td style="padding: 8px; border-bottom: 1px solid #eee;"><strong>状态:</strong></td>
                        <td style="padding: 8px; border-bottom: 1px solid #eee;">
                            <span class="provider-status status-${provider.status}">${this.getStatusText(provider.status)}</span>
                        </td></tr>
                    <tr><td style="padding: 8px; border-bottom: 1px solid #eee;"><strong>基础URL:</strong></td>
                        <td style="padding: 8px; border-bottom: 1px solid #eee;">${provider.baseUrl || 'N/A'}</td></tr>
                    <tr><td style="padding: 8px; border-bottom: 1px solid #eee;"><strong>超时设置:</strong></td>
                        <td style="padding: 8px; border-bottom: 1px solid #eee;">${provider.timeout}秒</td></tr>
                    <tr><td style="padding: 8px; border-bottom: 1px solid #eee;"><strong>重试次数:</strong></td>
                        <td style="padding: 8px; border-bottom: 1px solid #eee;">${provider.retries}</td></tr>
                    <tr><td style="padding: 8px; border-bottom: 1px solid #eee;"><strong>请求次数:</strong></td>
                        <td style="padding: 8px; border-bottom: 1px solid #eee;">${provider.requests || 0}</td></tr>
                    <tr><td style="padding: 8px; border-bottom: 1px solid #eee;"><strong>平均响应时间:</strong></td>
                        <td style="padding: 8px; border-bottom: 1px solid #eee;">${provider.avgResponseTime || 0}ms</td></tr>
                    <tr><td style="padding: 8px; border-bottom: 1px solid #eee;"><strong>Token消耗:</strong></td>
                        <td style="padding: 8px; border-bottom: 1px solid #eee;">${provider.tokens || 0}</td></tr>
                </table>
            </div>
            <div>
                <h4>支持的模型 (${models.length}个)</h4>
                <div style="display: flex; flex-wrap: wrap; gap: 8px; margin-top: 10px;">
                    ${models.map(model => `<span style="background: #e2e8f0; padding: 4px 8px; border-radius: 4px; font-size: 0.9rem;">${model}</span>`).join('')}
                </div>
            </div>
        `;

        this.showModal();
    }

    showCostStats(providerName) {
        const provider = this.providers.find(p => p.name === providerName);
        if (!provider) return;

        const modalTitle = document.getElementById('modal-title');
        const modalBody = document.getElementById('modal-body');

        modalTitle.textContent = `${provider.name.toUpperCase()} 成本统计`;
        
        // 根据不同提供商计算大概的成本
        const costInfo = this.calculateCost(provider.name, provider.tokens || 0);
        
        modalBody.innerHTML = `
            <div style="margin-bottom: 20px;">
                <h4><i class="fas fa-chart-pie"></i> 使用统计</h4>
                <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin: 16px 0;">
                    <div style="background: #f8fafc; padding: 16px; border-radius: 8px; text-align: center;">
                        <div style="font-size: 2rem; font-weight: bold; color: #667eea;">${provider.requests || 0}</div>
                        <div style="color: #666; margin-top: 4px;">总请求数</div>
                    </div>
                    <div style="background: #f8fafc; padding: 16px; border-radius: 8px; text-align: center;">
                        <div style="font-size: 2rem; font-weight: bold; color: #48bb78;">${provider.tokens || 0}</div>
                        <div style="color: #666; margin-top: 4px;">Token消耗</div>
                    </div>
                </div>
            </div>
            
            <div style="margin-bottom: 20px;">
                <h4><i class="fas fa-dollar-sign"></i> 成本估算</h4>
                <div style="background: #fff3cd; border: 1px solid #ffeaa7; border-radius: 8px; padding: 16px;">
                    <div style="margin-bottom: 12px;">
                        <strong>预估成本:</strong> 
                        <span style="font-size: 1.2rem; color: #e17055;">$${costInfo.estimatedCost}</span>
                    </div>
                    <div style="margin-bottom: 8px;">
                        <strong>定价模型:</strong> ${costInfo.pricingModel}
                    </div>
                    <div style="font-size: 0.9rem; color: #666;">
                        <i class="fas fa-info-circle"></i> 成本仅为估算值，实际费用请以各平台账单为准
                    </div>
                </div>
            </div>
            
            <div>
                <h4><i class="fas fa-clock"></i> 性能指标</h4>
                <table style="width: 100%; border-collapse: collapse;">
                    <tr><td style="padding: 8px; border-bottom: 1px solid #eee;"><strong>平均响应时间:</strong></td>
                        <td style="padding: 8px; border-bottom: 1px solid #eee;">${provider.avgResponseTime || 0}ms</td></tr>
                    <tr><td style="padding: 8px; border-bottom: 1px solid #eee;"><strong>平均Token/请求:</strong></td>
                        <td style="padding: 8px; border-bottom: 1px solid #eee;">${provider.requests > 0 ? Math.round((provider.tokens || 0) / provider.requests) : 0}</td></tr>
                    <tr><td style="padding: 8px; border-bottom: 1px solid #eee;"><strong>使用状态:</strong></td>
                        <td style="padding: 8px; border-bottom: 1px solid #eee;">
                            <span class="provider-status status-${provider.status}">${this.getStatusText(provider.status)}</span>
                        </td></tr>
                </table>
            </div>
        `;

        this.showModal();
    }

    calculateCost(providerName, tokens) {
        // 各提供商的大概定价（每1K tokens的美元成本）
        const pricingMap = {
            'openai': { input: 0.001, output: 0.002, model: 'GPT-3.5-turbo' },
            'gemini': { input: 0.0005, output: 0.0015, model: 'Gemini Pro' },
            'deepseek': { input: 0.0002, output: 0.0006, model: 'DeepSeek Chat' },
            'qwen': { input: 0.0008, output: 0.0024, model: '通义千问' },
            'moonshot': { input: 0.0024, output: 0.0072, model: 'Moonshot v1' }
        };

        const pricing = pricingMap[providerName] || { input: 0.001, output: 0.002, model: '未知模型' };
        
        // 假设输入输出比例为 1:1，实际使用中可以根据具体情况调整
        const inputTokens = tokens * 0.6;  // 假设60%为输入
        const outputTokens = tokens * 0.4; // 假设40%为输出
        
        const cost = (inputTokens / 1000 * pricing.input) + (outputTokens / 1000 * pricing.output);
        
        return {
            estimatedCost: cost.toFixed(4),
            pricingModel: pricing.model
        };
    }

    showModal() {
        document.getElementById('modal').style.display = 'block';
    }

    closeModal() {
        document.getElementById('modal').style.display = 'none';
    }

    showLoading() {
        document.getElementById('loading').style.display = 'block';
    }

    hideLoading() {
        document.getElementById('loading').style.display = 'none';
    }

    showError(message) {
        // 简单的错误提示，可以后续改进为toast通知
        alert('错误: ' + message);
    }

    updateLastRefreshTime() {
        const now = new Date();
        const timeString = now.toLocaleTimeString('zh-CN');
        document.getElementById('last-update').textContent = `最后更新: ${timeString}`;
    }

    getStatusText(status) {
        const statusMap = {
            'healthy': '健康',
            'unhealthy': '故障',
            'unknown': '未知'
        };
        return statusMap[status] || '未知';
    }

    startAutoRefresh() {
        this.refreshTimer = setInterval(() => {
            this.loadData();
        }, this.refreshInterval);
    }

    stopAutoRefresh() {
        if (this.refreshTimer) {
            clearInterval(this.refreshTimer);
            this.refreshTimer = null;
        }
    }

}

// 全局实例
let monitor;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    monitor = new LLMMonitor();
});

// 页面卸载前清理定时器
window.addEventListener('beforeunload', () => {
    if (monitor) {
        monitor.stopAutoRefresh();
    }
});