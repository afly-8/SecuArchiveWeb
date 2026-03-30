let currentUser = null;
let reports = [];
let projects = [];
let backups = [];
let categories = [];
let selectedFile = null;
let aiProviders = [];
let currentProviderModels = [];

const API_BASE = '/api';

async function apiRequest(url, options = {}) {
    const token = localStorage.getItem('token');
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    
    const response = await fetch(url, { ...options, headers });
    if (response.status === 401) {
        logout();
        return null;
    }
    return response;
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    window.location.href = '/login';
}

function checkAuth() {
    const token = localStorage.getItem('token');
    const user = localStorage.getItem('user');
    if (!token || !user) {
        window.location.href = '/login';
        return;
    }
    currentUser = JSON.parse(user);
    document.getElementById('userName').textContent = currentUser.nickname || currentUser.username;
}

document.addEventListener('DOMContentLoaded', () => {
    checkAuth();
    loadDashboard();
    setupNavigation();
    loadReportCategories();
    loadProjectsForSelect();
    loadAIConfig();
});

function setupNavigation() {
    document.querySelectorAll('.nav-link').forEach(link => {
        link.addEventListener('click', function() {
            const section = this.getAttribute('data-section');
            showSection(section);
        });
    });
}

function showSection(section) {
    document.querySelectorAll('.section').forEach(s => s.classList.remove('active'));
    document.querySelectorAll('.nav-link').forEach(l => l.classList.remove('active'));
    
    document.getElementById(section).classList.add('active');
    document.querySelector(`[data-section="${section}"]`).classList.add('active');
    
    const titles = {
        'dashboard': '仪表盘',
        'projects': '项目管理',
        'reports': '报告管理',
        'templates': '模板管理',
        'ai': 'AI助手',
        'backup': '备份管理',
        'settings': '系统设置'
    };
    document.getElementById('pageTitle').textContent = titles[section] || '仪表盘';
    
    switch(section) {
        case 'dashboard': loadDashboard(); break;
        case 'projects': loadProjects(); break;
        case 'reports': loadReports(); break;
        case 'templates': loadTemplates(); break;
        case 'ai': loadAIChat(); break;
        case 'backup': loadBackups(); break;
    }
}

async function loadDashboard() {
    const reportsRes = await apiRequest(`${API_BASE}/reports`);
    reports = reportsRes ? await reportsRes.json() : [];
    
    const projectsRes = await apiRequest(`${API_BASE}/projects`);
    projects = projectsRes ? await projectsRes.json() : [];
    
    const backupsRes = await apiRequest(`${API_BASE}/backups`);
    backups = backupsRes ? await backupsRes.json() : [];
    
    const aiRes = await apiRequest(`${API_BASE}/ai/config`);
    const aiConfig = aiRes ? await aiRes.json() : { is_enabled: false };
    
    document.getElementById('reportCount').textContent = reports.length;
    document.getElementById('projectCount').textContent = projects.length;
    document.getElementById('backupCount').textContent = backups.length;
    document.getElementById('aiStatus').textContent = aiConfig.is_enabled ? '已启用' : '未配置';
    
    const tbody = document.querySelector('#recentReportsTable tbody');
    tbody.innerHTML = '';
    reports.slice(0, 5).forEach(r => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${r.name}</td>
            <td><span class="badge-category">${r.category || '-'}</span></td>
            <td>${r.project ? r.project.name : '-'}</td>
            <td>${new Date(r.upload_time).toLocaleDateString()}</td>
        `;
        tbody.appendChild(tr);
    });
}

async function loadReports() {
    const res = await apiRequest(`${API_BASE}/reports`);
    if (res) {
        reports = await res.json();
        renderReports();
    }
}

function renderReports() {
    const tbody = document.querySelector('#reportsTable tbody');
    tbody.innerHTML = '';
    reports.forEach(r => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${r.name}</td>
            <td><span class="badge-category">${r.category || '-'}</span></td>
            <td>${r.sub_category || '-'}</td>
            <td>${r.project ? r.project.name : '-'}</td>
            <td>${formatFileSize(r.file_size)}</td>
            <td>${new Date(r.upload_time).toLocaleDateString()}</td>
            <td>
                <button class="btn-action btn-download" onclick="downloadReport(${r.id})" title="下载"><i class="fas fa-download"></i></button>
                <button class="btn-action btn-edit" onclick="editReport(${r.id})" title="编辑"><i class="fas fa-edit"></i></button>
                <button class="btn-action btn-delete" onclick="deleteReport(${r.id})" title="删除"><i class="fas fa-trash"></i></button>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

function formatFileSize(bytes) {
    if (!bytes) return '-';
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return (bytes / Math.pow(1024, i)).toFixed(2) + ' ' + sizes[i];
}

function searchReports() {
    const keyword = document.getElementById('reportSearch').value.toLowerCase();
    const filtered = reports.filter(r => 
        r.name.toLowerCase().includes(keyword) || 
        (r.tags && r.tags.toLowerCase().includes(keyword))
    );
    const tbody = document.querySelector('#reportsTable tbody');
    tbody.innerHTML = '';
    filtered.forEach(r => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${r.name}</td>
            <td><span class="badge-category">${r.category || '-'}</span></td>
            <td>${r.sub_category || '-'}</td>
            <td>${r.project ? r.project.name : '-'}</td>
            <td>${formatFileSize(r.file_size)}</td>
            <td>${new Date(r.upload_time).toLocaleDateString()}</td>
            <td>
                <button class="btn-action btn-download" onclick="downloadReport(${r.id})"><i class="fas fa-download"></i></button>
                <button class="btn-action btn-edit" onclick="editReport(${r.id})"><i class="fas fa-edit"></i></button>
                <button class="btn-action btn-delete" onclick="deleteReport(${r.id})"><i class="fas fa-trash"></i></button>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

function filterReports() {
    const category = document.getElementById('reportCategoryFilter').value;
    const keyword = document.getElementById('reportSearch').value.toLowerCase();
    
    let filtered = reports;
    if (category) {
        filtered = filtered.filter(r => r.category === category);
    }
    if (keyword) {
        filtered = filtered.filter(r => 
            r.name.toLowerCase().includes(keyword) || 
            (r.tags && r.tags.toLowerCase().includes(keyword))
        );
    }
    
    const tbody = document.querySelector('#reportsTable tbody');
    tbody.innerHTML = '';
    filtered.forEach(r => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${r.name}</td>
            <td><span class="badge-category">${r.category || '-'}</span></td>
            <td>${r.sub_category || '-'}</td>
            <td>${r.project ? r.project.name : '-'}</td>
            <td>${formatFileSize(r.file_size)}</td>
            <td>${new Date(r.upload_time).toLocaleDateString()}</td>
            <td>
                <button class="btn-action btn-download" onclick="downloadReport(${r.id})"><i class="fas fa-download"></i></button>
                <button class="btn-action btn-edit" onclick="editReport(${r.id})"><i class="fas fa-edit"></i></button>
                <button class="btn-action btn-delete" onclick="deleteReport(${r.id})"><i class="fas fa-trash"></i></button>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

async function loadReportCategories() {
    const res = await apiRequest(`${API_BASE}/reports/categories`);
    if (res) {
        categories = await res.json();
        
        const catSelect = document.getElementById('reportCategory');
        catSelect.innerHTML = '<option value="">请选择分类</option>';
        categories.forEach(c => {
            const opt = document.createElement('option');
            opt.value = c.name;
            opt.textContent = c.name;
            catSelect.appendChild(opt);
        });
        
        const filterSelect = document.getElementById('reportCategoryFilter');
        filterSelect.innerHTML = '<option value="">全部分类</option>';
        categories.forEach(c => {
            const opt = document.createElement('option');
            opt.value = c.name;
            opt.textContent = c.name;
            filterSelect.appendChild(opt);
        });
    }
}

function updateSubCategories() {
    const category = document.getElementById('reportCategory').value;
    const subCatSelect = document.getElementById('reportSubCategory');
    subCatSelect.innerHTML = '';
    
    const cat = categories.find(c => c.name === category);
    if (cat) {
        cat.sub_categories.forEach(sc => {
            const opt = document.createElement('option');
            opt.value = sc;
            opt.textContent = sc;
            subCatSelect.appendChild(opt);
        });
    }
}

function showImportModal() {
    selectedFile = null;
    document.getElementById('selectedFileName').textContent = '';
    document.getElementById('reportName').value = '';
    document.getElementById('reportTags').value = '';
    document.getElementById('reportDescription').value = '';
    new bootstrap.Modal(document.getElementById('importModal')).show();
}

function handleFileSelect(input) {
    if (input.files && input.files[0]) {
        selectedFile = input.files[0];
        document.getElementById('selectedFileName').textContent = selectedFile.name;
        document.getElementById('reportName').value = selectedFile.name;
    }
}

async function importReport() {
    if (!selectedFile) {
        alert('请选择文件');
        return;
    }
    
    const formData = new FormData();
    formData.append('file', selectedFile);
    formData.append('name', document.getElementById('reportName').value);
    formData.append('project_id', document.getElementById('reportProject').value);
    formData.append('category', document.getElementById('reportCategory').value);
    formData.append('sub_category', document.getElementById('reportSubCategory').value);
    formData.append('tags', document.getElementById('reportTags').value);
    formData.append('description', document.getElementById('reportDescription').value);
    
    const token = localStorage.getItem('token');
    const res = await fetch(`${API_BASE}/reports`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
        body: formData
    });
    
    if (res.ok) {
        bootstrap.Modal.getInstance(document.getElementById('importModal')).hide();
        loadReports();
        alert('报告导入成功');
    } else {
        alert('导入失败');
    }
}

async function downloadReport(id) {
    const token = localStorage.getItem('token');
    window.open(`${API_BASE}/reports/${id}/download?token=${token}`, '_blank');
}

async function deleteReport(id) {
    if (!confirm('确定要删除这个报告吗?')) return;
    
    const res = await apiRequest(`${API_BASE}/reports/${id}`, { method: 'DELETE' });
    if (res && res.ok) {
        loadReports();
        alert('删除成功');
    }
}

async function editReport(id) {
    const report = reports.find(r => r.id === id);
    if (!report) return;
    
    document.getElementById('reportName').value = report.name;
    document.getElementById('reportCategory').value = report.category || '';
    updateSubCategories();
    document.getElementById('reportSubCategory').value = report.sub_category || '';
    document.getElementById('reportTags').value = report.tags || '';
    document.getElementById('reportDescription').value = report.description || '';
    
    const projectSelect = document.getElementById('reportProject');
    if (report.project_id) {
        projectSelect.value = report.project_id;
    }
    
    selectedFile = { name: report.file_name };
    document.getElementById('selectedFileName').textContent = report.file_name;
    
    const formData = new FormData();
    formData.append('name', report.name);
    formData.append('project_id', report.project_id || '');
    formData.append('category', report.category || '');
    formData.append('sub_category', report.sub_category || '');
    formData.append('tags', report.tags || '');
    formData.append('description', report.description || '');
    
    const token = localStorage.getItem('token');
    await fetch(`${API_BASE}/reports/${id}`, {
        method: 'PUT',
        headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
        body: JSON.stringify({
            name: report.name,
            project_id: report.project_id,
            category: report.category,
            sub_category: report.sub_category,
            tags: report.tags,
            description: report.description
        })
    });
    
    bootstrap.Modal.getInstance(document.getElementById('importModal')).show();
    loadReports();
}

async function loadProjects() {
    const res = await apiRequest(`${API_BASE}/projects`);
    if (res) {
        projects = await res.json();
        renderProjects();
    }
}

function renderProjects() {
    const tbody = document.querySelector('#projectsTable tbody');
    tbody.innerHTML = '';
    projects.forEach(p => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${p.name}</td>
            <td>${p.category || '-'}</td>
            <td>${p.client_name || '-'}</td>
            <td><span class="badge-category">${p.status === 'active' ? '进行中' : '已完成'}</span></td>
            <td>${p.reports ? p.reports.length : 0}</td>
            <td>${new Date(p.created_at).toLocaleDateString()}</td>
            <td>
                <button class="btn-action btn-edit" onclick="editProject(${p.id})"><i class="fas fa-edit"></i></button>
                <button class="btn-action btn-delete" onclick="deleteProject(${p.id})"><i class="fas fa-trash"></i></button>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

async function loadProjectsForSelect() {
    const res = await apiRequest(`${API_BASE}/projects`);
    if (res) {
        const projectList = await res.json();
        const select = document.getElementById('reportProject');
        select.innerHTML = '<option value="">无</option>';
        projectList.forEach(p => {
            const opt = document.createElement('option');
            opt.value = p.id;
            opt.textContent = p.name;
            select.appendChild(opt);
        });
    }
}

function showProjectModal(id = null) {
    document.getElementById('projectId').value = id || '';
    document.getElementById('projectModalTitle').textContent = id ? '编辑项目' : '创建项目';
    document.getElementById('projectName').value = '';
    document.getElementById('projectCategory').value = '渗透测试';
    document.getElementById('projectClient').value = '';
    document.getElementById('projectDescription').value = '';
    
    if (id) {
        const p = projects.find(pr => pr.id === id);
        if (p) {
            document.getElementById('projectName').value = p.name;
            document.getElementById('projectCategory').value = p.category || '渗透测试';
            document.getElementById('projectClient').value = p.client_name || '';
            document.getElementById('projectDescription').value = p.description || '';
        }
    }
    
    new bootstrap.Modal(document.getElementById('projectModal')).show();
}

function editProject(id) {
    showProjectModal(id);
}

async function saveProject() {
    const id = document.getElementById('projectId').value;
    const zipFile = document.getElementById('projectZipFile').files[0];
    
    if (!document.getElementById('projectName').value) {
        alert('请输入项目名称');
        return;
    }
    
    if (id) {
        const data = {
            name: document.getElementById('projectName').value,
            category: document.getElementById('projectCategory').value,
            client_name: document.getElementById('projectClient').value,
            contract: document.getElementById('projectContract').value,
            contract_no: document.getElementById('projectContractNo').value,
            description: document.getElementById('projectDescription').value
        };
        
        const res = await apiRequest(`${API_BASE}/projects/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        
        if (res && res.ok) {
            bootstrap.Modal.getInstance(document.getElementById('projectModal')).hide();
            loadProjects();
        }
    } else if (zipFile) {
        const formData = new FormData();
        formData.append('name', document.getElementById('projectName').value);
        formData.append('category', document.getElementById('projectCategory').value);
        formData.append('client_name', document.getElementById('projectClient').value);
        formData.append('contract', document.getElementById('projectContract').value);
        formData.append('contract_no', document.getElementById('projectContractNo').value);
        formData.append('description', document.getElementById('projectDescription').value);
        formData.append('file', zipFile);
        
        const token = localStorage.getItem('token');
        const res = await fetch(`${API_BASE}/projects/import-zip`, {
            method: 'POST',
            headers: { 'Authorization': `Bearer ${token}` },
            body: formData
        });
        
        if (res.ok) {
            const result = await res.json();
            alert(`项目创建成功，已自动导入 ${result.reportCount} 份报告`);
            bootstrap.Modal.getInstance(document.getElementById('projectModal')).hide();
            loadProjects();
            loadProjectsForSelect();
        } else {
            alert('创建项目失败');
        }
    } else {
        const data = {
            name: document.getElementById('projectName').value,
            category: document.getElementById('projectCategory').value,
            client_name: document.getElementById('projectClient').value,
            contract: document.getElementById('projectContract').value,
            contract_no: document.getElementById('projectContractNo').value,
            description: document.getElementById('projectDescription').value
        };
        
        const res = await apiRequest(`${API_BASE}/projects`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(data)
        });
        
        if (res && res.ok) {
            bootstrap.Modal.getInstance(document.getElementById('projectModal')).hide();
            loadProjects();
            loadProjectsForSelect();
        }
    }
}

async function deleteProject(id) {
    if (!confirm('确定要删除这个项目吗?')) return;
    
    const res = await apiRequest(`${API_BASE}/projects/${id}`, { method: 'DELETE' });
    if (res && res.ok) {
        loadProjects();
        loadProjectsForSelect();
    }
}

function handleChatKeypress(e) {
    if (e.key === 'Enter') {
        sendChat();
    }
}

async function sendChat() {
    const input = document.getElementById('chatInput');
    const message = input.value.trim();
    if (!message) return;
    
    const container = document.getElementById('chatContainer');
    
    const userMsg = document.createElement('div');
    userMsg.className = 'chat-message user';
    userMsg.textContent = message;
    container.appendChild(userMsg);
    
    input.value = '';
    
    const loading = document.createElement('div');
    loading.className = 'chat-message ai';
    loading.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 思考中...';
    container.appendChild(loading);
    container.scrollTop = container.scrollHeight;
    
    const res = await apiRequest(`${API_BASE}/ai/chat-with-context`, {
        method: 'POST',
        body: JSON.stringify({ message })
    });
    
    loading.remove();
    
    if (res) {
        const data = await res.json();
        const aiMsg = document.createElement('div');
        aiMsg.className = 'chat-message ai';
        aiMsg.textContent = data.response || '无响应';
        container.appendChild(aiMsg);
    } else {
        const errMsg = document.createElement('div');
        errMsg.className = 'chat-message ai';
        errMsg.textContent = 'AI 服务未配置或已禁用';
        container.appendChild(errMsg);
    }
    
    container.scrollTop = container.scrollHeight;
}

function loadAIChat() {
    const container = document.getElementById('chatContainer');
    if (container.children.length === 0) {
        const welcome = document.createElement('div');
        welcome.className = 'chat-message ai';
        welcome.textContent = '你好！我是 AI 助手，可以帮你分析和查询安全服务报告。有什么问题尽管问我！';
        container.appendChild(welcome);
    }
}

async function loadBackups() {
    const res = await apiRequest(`${API_BASE}/backups`);
    if (res) {
        backups = await res.json();
        renderBackups();
    }
}

let templates = [];

async function loadTemplates() {
    const res = await apiRequest(`${API_BASE}/templates`);
    if (res) {
        templates = await res.json();
        renderTemplates();
    }
}

function renderTemplates() {
    const tbody = document.querySelector('#templatesTable tbody');
    tbody.innerHTML = '';
    templates.forEach(t => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${t.name}</td>
            <td><span class="badge-category">${t.category || '-'}</span></td>
            <td>${t.file_name}</td>
            <td>${formatFileSize(t.file_size)}</td>
            <td>${new Date(t.created_at).toLocaleDateString()}</td>
            <td>
                <button class="btn-action btn-download" onclick="downloadTemplate(${t.id})" title="下载"><i class="fas fa-download"></i></button>
                <button class="btn-action btn-delete" onclick="deleteTemplate(${t.id})" title="删除"><i class="fas fa-trash"></i></button>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

function showImportTemplateModal() {
    document.getElementById('templateName').value = '';
    document.getElementById('templateDescription').value = '';
    document.getElementById('templateFile').value = '';
    new bootstrap.Modal(document.getElementById('templateModal')).show();
}

async function importTemplate() {
    const name = document.getElementById('templateName').value;
    const file = document.getElementById('templateFile').files[0];
    
    if (!name || !file) {
        alert('请填写模板名称并选择文件');
        return;
    }
    
    const formData = new FormData();
    formData.append('name', name);
    formData.append('category', document.getElementById('templateCategory').value);
    formData.append('description', document.getElementById('templateDescription').value);
    formData.append('file', file);
    
    const token = localStorage.getItem('token');
    const res = await fetch(`${API_BASE}/templates`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
        body: formData
    });
    
    if (res.ok) {
        bootstrap.Modal.getInstance(document.getElementById('templateModal')).hide();
        loadTemplates();
        alert('模板导入成功');
    } else {
        alert('模板导入失败');
    }
}

async function downloadTemplate(id) {
    const token = localStorage.getItem('token');
    window.open(`${API_BASE}/templates/${id}/download?token=${token}`, '_blank');
}

async function deleteTemplate(id) {
    if (!confirm('确定要删除这个模板吗?')) return;
    
    const res = await apiRequest(`${API_BASE}/templates/${id}`, { method: 'DELETE' });
    if (res && res.ok) {
        loadTemplates();
    }
}

function renderBackups() {
    const tbody = document.querySelector('#backupsTable tbody');
    tbody.innerHTML = '';
    backups.forEach(b => {
        const tr = document.createElement('tr');
        tr.innerHTML = `
            <td>${b.file_name}</td>
            <td>${formatFileSize(b.file_size)}</td>
            <td><span class="badge-category">${b.type}</span></td>
            <td>${new Date(b.created_at).toLocaleDateString()}</td>
            <td>
                <button class="btn-action btn-download" onclick="downloadBackup(${b.id})"><i class="fas fa-download"></i></button>
                <button class="btn-action btn-edit" onclick="restoreBackup(${b.id})"><i class="fas fa-redo"></i></button>
                <button class="btn-action btn-delete" onclick="deleteBackup(${b.id})"><i class="fas fa-trash"></i></button>
            </td>
        `;
        tbody.appendChild(tr);
    });
}

async function createBackup() {
    const description = prompt('请输入备份描述:');
    const res = await apiRequest(`${API_BASE}/backups`, {
        method: 'POST',
        body: JSON.stringify({ description: description || '' })
    });
    
    if (res && res.ok) {
        loadBackups();
        alert('备份创建成功');
    }
}

async function importBackup(input) {
    if (!input.files || !input.files[0]) return;
    
    const formData = new FormData();
    formData.append('file', input.files[0]);
    
    const token = localStorage.getItem('token');
    const res = await fetch(`${API_BASE}/backups/import`, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
        body: formData
    });
    
    if (res.ok) {
        loadBackups();
        loadDashboard();
        alert('备份导入成功');
    } else {
        alert('导入失败');
    }
}

async function downloadBackup(id) {
    const token = localStorage.getItem('token');
    window.open(`${API_BASE}/backups/${id}/download?token=${token}`, '_blank');
}

async function restoreBackup(id) {
    if (!confirm('确定要恢复这个备份吗? 当前数据将被覆盖。')) return;
    
    const res = await apiRequest(`${API_BASE}/backups/${id}/restore`, { method: 'POST' });
    if (res && res.ok) {
        loadDashboard();
        loadReports();
        loadProjects();
        alert('备份恢复成功');
    }
}

async function deleteBackup(id) {
    if (!confirm('确定要删除这个备份吗?')) return;
    
    const res = await apiRequest(`${API_BASE}/backups/${id}`, { method: 'DELETE' });
    if (res && res.ok) {
        loadBackups();
    }
}

async function loadAIProviders() {
    const res = await apiRequest(`${API_BASE}/ai/providers`);
    if (res) {
        aiProviders = await res.json();
        renderAIProviders();
    }
}

function renderAIProviders() {
    const select = document.getElementById('aiProvider');
    select.innerHTML = '';
    aiProviders.forEach(p => {
        const opt = document.createElement('option');
        opt.value = p.id;
        opt.textContent = p.name;
        if (p.id === 'custom') {
            opt.textContent = p.name + ' (自定义)';
        }
        select.appendChild(opt);
    });
}

function onProviderChange() {
    const providerId = document.getElementById('aiProvider').value;
    const provider = aiProviders.find(p => p.id === providerId);
    const modelSelect = document.getElementById('aiModel');
    const customModelSection = document.getElementById('customModelSection');
    
    modelSelect.innerHTML = '';
    
    if (provider && provider.models && provider.models.length > 0) {
        currentProviderModels = provider.models;
        provider.models.forEach(m => {
            const opt = document.createElement('option');
            opt.value = m;
            opt.textContent = m;
            modelSelect.appendChild(opt);
        });
        document.getElementById('aiUseCustomModel').checked = false;
        customModelSection.style.display = 'none';
    } else if (providerId === 'custom') {
        document.getElementById('aiUseCustomModel').checked = true;
        customModelSection.style.display = 'block';
        const opt = document.createElement('option');
        opt.value = '';
        opt.textContent = '请输入自定义模型名称';
        modelSelect.appendChild(opt);
    } else if (providerId === 'ollama') {
        const opt = document.createElement('option');
        opt.value = '';
        opt.textContent = '请在API端点输入本地地址';
        modelSelect.appendChild(opt);
        document.getElementById('aiUseCustomModel').checked = false;
        customModelSection.style.display = 'none';
    } else {
        const opt = document.createElement('option');
        opt.value = '';
        opt.textContent = '无可用模型';
        modelSelect.appendChild(opt);
        document.getElementById('aiUseCustomModel').checked = false;
        customModelSection.style.display = 'none';
    }
    
    if (provider && provider.endpoint && document.getElementById('aiEndpoint').value === '') {
        document.getElementById('aiEndpoint').placeholder = provider.endpoint;
    }
}

function toggleCustomModel() {
    const useCustom = document.getElementById('aiUseCustomModel').checked;
    const customModelSection = document.getElementById('customModelSection');
    const modelSelect = document.getElementById('aiModel');
    
    if (useCustom) {
        customModelSection.style.display = 'block';
        modelSelect.innerHTML = '<option value="">自定义模型</option>';
    } else {
        customModelSection.style.display = 'none';
        onProviderChange();
    }
}

async function showAIConfigModal() {
    await loadAIProviders();
    await loadAIConfig();
    new bootstrap.Modal(document.getElementById('aiConfigModal')).show();
}

async function loadAIConfig() {
    const res = await apiRequest(`${API_BASE}/ai/config`);
    if (res) {
        const config = await res.json();
        
        const providerSelect = document.getElementById('aiProvider');
        if (config.provider) {
            providerSelect.value = config.provider;
            onProviderChange();
        }
        
        if (config.custom_model && config.custom_model !== '') {
            document.getElementById('aiUseCustomModel').checked = true;
            document.getElementById('customModelSection').style.display = 'block';
            document.getElementById('aiCustomModel').value = config.custom_model;
            
            const modelSelect = document.getElementById('aiModel');
            const opt = document.createElement('option');
            opt.value = config.custom_model;
            opt.textContent = config.custom_model + ' (自定义)';
            modelSelect.innerHTML = '';
            modelSelect.appendChild(opt);
            modelSelect.value = config.custom_model;
        } else if (config.model) {
            const modelSelect = document.getElementById('aiModel');
            if (currentProviderModels.includes(config.model)) {
                modelSelect.value = config.model;
            } else {
                const opt = document.createElement('option');
                opt.value = config.model;
                opt.textContent = config.model;
                modelSelect.appendChild(opt);
                modelSelect.value = config.model;
            }
        }
        
        document.getElementById('aiEndpoint').value = config.api_endpoint || '';
        document.getElementById('aiEnabled').checked = config.is_enabled || false;
    }
}

async function saveAIConfig() {
    const useCustom = document.getElementById('aiUseCustomModel').checked;
    const customModel = document.getElementById('aiCustomModel').value;
    const selectedModel = document.getElementById('aiModel').value;
    
    const data = {
        provider: document.getElementById('aiProvider').value,
        api_key: document.getElementById('aiApiKey').value,
        api_endpoint: document.getElementById('aiEndpoint').value,
        model: useCustom ? customModel : selectedModel,
        custom_model: useCustom ? customModel : '',
        is_enabled: document.getElementById('aiEnabled').checked
    };
    
    if (!data.provider) {
        alert('请选择AI提供商');
        return;
    }
    
    if (useCustom && !customModel) {
        alert('请输入自定义模型名称');
        return;
    }
    
    const res = await apiRequest(`${API_BASE}/ai/config`, {
        method: 'POST',
        body: JSON.stringify(data)
    });
    
    if (res && res.ok) {
        alert('配置保存成功');
        bootstrap.Modal.getInstance(document.getElementById('aiConfigModal')).hide();
    } else {
        alert('配置保存失败');
    }
}

async function testAIConnection() {
    const res = await apiRequest(`${API_BASE}/ai/test`, { method: 'POST' });
    if (res && res.ok) {
        alert('连接成功');
    } else {
        alert('连接失败');
    }
}

async function updatePassword() {
    const newPassword = document.getElementById('newPassword').value;
    if (!newPassword) {
        alert('请输入新密码');
        return;
    }
    
    const res = await apiRequest(`${API_BASE}/auth/user`, {
        method: 'PUT',
        body: JSON.stringify({ password: newPassword })
    });
    
    if (res && res.ok) {
        alert('密码修改成功，请重新登录');
        logout();
    }
}
