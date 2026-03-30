package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"

	"secuarchive-web/internal/config"
	"secuarchive-web/internal/handlers"
	"secuarchive-web/internal/middleware"
	"secuarchive-web/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Init()

	err := services.InitDB()
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		return
	}

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())

	r.SetHTMLTemplate(loadTemplates())

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", nil)
	})

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login", nil)
	})

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/register", handlers.Register)
			auth.GET("/me", middleware.AuthMiddleware(), handlers.GetCurrentUser)
			auth.PUT("/user", middleware.AuthMiddleware(), handlers.UpdateUser)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/users", handlers.GetUsers)

			projects := protected.Group("/projects")
			{
				projects.GET("", handlers.GetProjects)
				projects.GET("/categories", handlers.GetProjectCategories)
				projects.GET("/:id", handlers.GetProject)
				projects.POST("", handlers.CreateProject)
				projects.POST("/import-zip", handlers.CreateProjectWithZip)
				projects.PUT("/:id", handlers.UpdateProject)
				projects.DELETE("/:id", handlers.DeleteProject)
			}

			reports := protected.Group("/reports")
			{
				reports.GET("", handlers.GetReports)
				reports.GET("/categories", handlers.GetReportCategories)
				reports.GET("/:id", handlers.GetReport)
				reports.POST("", handlers.ImportReport)
				reports.PUT("/:id", handlers.UpdateReport)
				reports.DELETE("/:id", handlers.DeleteReport)
				reports.GET("/:id/download", handlers.DownloadReport)
				reports.GET("/:id/preview", handlers.PreviewReport)
			}

			backups := protected.Group("/backups")
			{
				backups.GET("", handlers.GetBackups)
				backups.POST("", handlers.CreateBackup)
				backups.POST("/import", handlers.ImportBackup)
				backups.GET("/:id/download", handlers.DownloadBackup)
				backups.POST("/:id/restore", handlers.RestoreBackup)
				backups.DELETE("/:id", handlers.DeleteBackup)
			}

			ai := protected.Group("/ai")
			{
				ai.GET("/config", handlers.GetAIConfig)
				ai.POST("/config", handlers.SaveAIConfig)
				ai.GET("/providers", handlers.GetAIProviders)
				ai.POST("/chat", handlers.AIChat)
				ai.POST("/chat-with-context", handlers.ChatWithContext)
				ai.POST("/test", handlers.TestAIConnection)
			}

			templates := protected.Group("/templates")
			{
				templates.GET("", handlers.GetTemplates)
				templates.GET("/categories", handlers.GetTemplateCategories)
				templates.GET("/:id", handlers.GetTemplate)
				templates.POST("", handlers.ImportTemplate)
				templates.GET("/:id/download", handlers.DownloadTemplate)
				templates.DELETE("/:id", handlers.DeleteTemplate)
			}
		}
	}

	staticDir := filepath.Join(".", "web")
	r.Static("/static", staticDir)

	port := config.AppConfig.ServerPort
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("Open http://localhost:%s in your browser\n", port)
	r.Run(":" + port)
}

func loadTemplates() *template.Template {
	tmpl := template.New("")

	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SecuArchive - 安全服务报告归档系统</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root {
            --primary-color: #1e3a5f;
            --accent-color: #00d4aa;
            --bg-color: #1a1a2e;
            --card-bg: #16213e;
            --text-color: #e8e8e8;
            --border-color: #2a2a4a;
        }
        
        * {
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-color);
            margin: 0;
            padding: 0;
            min-height: 100vh;
        }
        
        .login-container {
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            background: linear-gradient(135deg, var(--bg-color) 0%, var(--primary-color) 100%);
        }
        
        .login-card {
            background: var(--card-bg);
            border-radius: 16px;
            padding: 40px;
            width: 100%;
            max-width: 400px;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
            border: 1px solid var(--border-color);
        }
        
        .login-logo {
            text-align: center;
            margin-bottom: 30px;
        }
        
        .login-logo i {
            font-size: 48px;
            color: var(--accent-color);
        }
        
        .login-logo h2 {
            margin-top: 10px;
            color: var(--text-color);
        }
        
        .form-control {
            background: var(--bg-color);
            border: 1px solid var(--border-color);
            color: var(--text-color);
            padding: 12px 15px;
            border-radius: 8px;
        }
        
        .form-control:focus {
            background: var(--bg-color);
            border-color: var(--accent-color);
            color: var(--text-color);
            box-shadow: 0 0 0 0.2rem rgba(0, 212, 170, 0.25);
        }
        
        .form-control::placeholder {
            color: #888;
        }
        
        .btn-primary {
            background: var(--accent-color);
            border: none;
            padding: 12px 24px;
            border-radius: 8px;
            font-weight: 600;
            width: 100%;
            transition: all 0.3s;
        }
        
        .btn-primary:hover {
            background: #00b894;
            transform: translateY(-2px);
            box-shadow: 0 5px 20px rgba(0, 212, 170, 0.3);
        }
        
        .sidebar {
            background: var(--card-bg);
            min-height: 100vh;
            border-right: 1px solid var(--border-color);
            padding: 20px 0;
        }
        
        .sidebar .nav-link {
            color: var(--text-color);
            padding: 12px 20px;
            margin: 5px 10px;
            border-radius: 8px;
            transition: all 0.3s;
        }
        
        .sidebar .nav-link:hover, .sidebar .nav-link.active {
            background: var(--accent-color);
            color: var(--bg-color);
        }
        
        .sidebar .nav-link i {
            margin-right: 10px;
            width: 20px;
        }
        
        .main-content {
            padding: 20px;
            background: var(--bg-color);
            min-height: 100vh;
        }
        
        .top-bar {
            background: var(--card-bg);
            padding: 15px 20px;
            border-radius: 12px;
            margin-bottom: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            border: 1px solid var(--border-color);
        }
        
        .card {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 12px;
            margin-bottom: 20px;
        }
        
        .card-header {
            background: transparent;
            border-bottom: 1px solid var(--border-color);
            padding: 15px 20px;
            font-weight: 600;
        }
        
        .card-body {
            padding: 20px;
        }
        
        .table {
            color: var(--text-color);
        }
        
        .table thead th {
            border-bottom: 1px solid var(--border-color);
            color: var(--accent-color);
            font-weight: 600;
        }
        
        .table td {
            border-bottom: 1px solid var(--border-color);
            vertical-align: middle;
        }
        
        .badge-category {
            background: var(--accent-color);
            color: var(--bg-color);
            padding: 5px 10px;
            border-radius: 15px;
            font-size: 12px;
        }
        
        .btn-action {
            padding: 5px 10px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            transition: all 0.3s;
            margin: 0 2px;
        }
        
        .btn-action.btn-edit {
            background: #3498db;
            color: white;
        }
        
        .btn-action.btn-delete {
            background: #e74c3c;
            color: white;
        }
        
        .btn-action.btn-download {
            background: #2ecc71;
            color: white;
        }
        
        .btn-action:hover {
            transform: scale(1.1);
        }
        
        .search-box {
            background: var(--bg-color);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 10px 15px;
            color: var(--text-color);
            width: 300px;
        }
        
        .modal-content {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 16px;
        }
        
        .modal-header {
            border-bottom: 1px solid var(--border-color);
        }
        
        .modal-footer {
            border-top: 1px solid var(--border-color);
        }
        
        .modal-body {
            padding: 20px;
        }
        
        .form-label {
            color: var(--text-color);
            margin-bottom: 8px;
        }
        
        .btn-close {
            filter: invert(1);
        }
        
        .nav-tabs {
            border-bottom: 1px solid var(--border-color);
        }
        
        .nav-tabs .nav-link {
            color: var(--text-color);
            border: none;
            padding: 12px 20px;
        }
        
        .nav-tabs .nav-link.active {
            background: var(--accent-color);
            color: var(--bg-color);
            border-radius: 8px 8px 0 0;
        }
        
        .nav-tabs .nav-link:hover {
            border: none;
        }
        
        .ai-chat-container {
            height: 400px;
            overflow-y: auto;
            background: var(--bg-color);
            border-radius: 8px;
            padding: 15px;
            margin-bottom: 15px;
        }
        
        .chat-message {
            margin-bottom: 15px;
            padding: 10px 15px;
            border-radius: 10px;
            max-width: 80%;
        }
        
        .chat-message.user {
            background: var(--accent-color);
            color: var(--bg-color);
            margin-left: auto;
        }
        
        .chat-message.ai {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
        }
        
        .file-drop-zone {
            border: 2px dashed var(--border-color);
            border-radius: 12px;
            padding: 40px;
            text-align: center;
            cursor: pointer;
            transition: all 0.3s;
        }
        
        .file-drop-zone:hover {
            border-color: var(--accent-color);
            background: rgba(0, 212, 170, 0.1);
        }
        
        .file-drop-zone i {
            font-size: 48px;
            color: var(--accent-color);
            margin-bottom: 15px;
        }
        
        .stats-card {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 12px;
            padding: 20px;
            text-align: center;
        }
        
        .stats-card h3 {
            font-size: 32px;
            color: var(--accent-color);
            margin: 10px 0;
        }
        
        .stats-card p {
            color: #888;
            margin: 0;
        }
        
        .splitter {
            width: 5px;
            background: var(--border-color);
            cursor: col-resize;
            transition: background 0.3s;
        }
        
        .splitter:hover {
            background: var(--accent-color);
        }
        
        .loading {
            display: none;
            text-align: center;
            padding: 20px;
        }
        
        .loading i {
            font-size: 32px;
            color: var(--accent-color);
            animation: spin 1s linear infinite;
        }
        
        @keyframes spin {
            100% { transform: rotate(360deg); }
        }
        
        .alert {
            border-radius: 8px;
            border: none;
        }
        
        .toast-container {
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 9999;
        }
        
        ::-webkit-scrollbar {
            width: 8px;
            height: 8px;
        }
        
        ::-webkit-scrollbar-track {
            background: var(--bg-color);
        }
        
        ::-webkit-scrollbar-thumb {
            background: var(--border-color);
            border-radius: 4px;
        }
        
        ::-webkit-scrollbar-thumb:hover {
            background: var(--accent-color);
        }
    </style>
</head>
<body>
    {{template "content" .}}
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/js/all.min.js"></script>
    <script src="/static/js/app.js"></script>
</body>
</html>
`

	tmpl, err := tmpl.Parse(html)
	if err != nil {
		panic(err)
	}

	loginTmpl, err := tmpl.Parse(loginHTML)
	if err != nil {
		panic(err)
	}
	indexTmpl, err := loginTmpl.Parse(indexHTML)
	if err != nil {
		panic(err)
	}

	return indexTmpl
}

const loginHTML = `{{define "login"}}
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>登录 - SecuArchive</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root {
            --primary-color: #1e3a5f;
            --accent-color: #00d4aa;
            --bg-color: #1a1a2e;
            --card-bg: #16213e;
            --text-color: #e8e8e8;
            --border-color: #2a2a4a;
        }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, var(--bg-color) 0%, var(--primary-color) 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0;
            padding: 20px;
        }
        .login-card {
            background: var(--card-bg);
            border-radius: 16px;
            padding: 40px;
            width: 100%;
            max-width: 400px;
            box-shadow: 0 10px 40px rgba(0, 0, 0, 0.3);
            border: 1px solid var(--border-color);
        }
        .login-logo {
            text-align: center;
            margin-bottom: 30px;
        }
        .login-logo i {
            font-size: 48px;
            color: var(--accent-color);
        }
        .login-logo h2 {
            margin-top: 10px;
            color: var(--text-color);
        }
        .form-control {
            background: var(--bg-color);
            border: 1px solid var(--border-color);
            color: var(--text-color);
            padding: 12px 15px;
            border-radius: 8px;
        }
        .form-control:focus {
            background: var(--bg-color);
            border-color: var(--accent-color);
            color: var(--text-color);
            box-shadow: 0 0 0 0.2rem rgba(0, 212, 170, 0.25);
        }
        .form-control::placeholder {
            color: #888;
        }
        .btn-primary {
            background: var(--accent-color);
            border: none;
            padding: 12px 24px;
            border-radius: 8px;
            font-weight: 600;
            width: 100%;
            color: var(--bg-color);
            transition: all 0.3s;
        }
        .btn-primary:hover {
            background: #00b894;
            transform: translateY(-2px);
            box-shadow: 0 5px 20px rgba(0, 212, 170, 0.3);
        }
    </style>
</head>
<body>
    <div class="login-card">
        <div class="login-logo">
            <i class="fas fa-shield-halved"></i>
            <h2>SecuArchive</h2>
            <p style="color: #888;">安全服务报告归档系统</p>
        </div>
        <form id="loginForm">
            <div class="mb-3">
                <label class="form-label">用户名</label>
                <input type="text" class="form-control" id="username" placeholder="请输入用户名" required>
            </div>
            <div class="mb-3">
                <label class="form-label">密码</label>
                <input type="password" class="form-control" id="password" placeholder="请输入密码" required>
            </div>
            <button type="submit" class="btn btn-primary">
                <i class="fas fa-sign-in-alt me-2"></i>登录
            </button>
        </form>
        <div class="mt-3 text-center">
            <small style="color: #888;">默认账号: admin / admin123</small>
        </div>
    </div>
    <script>
        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            
            try {
                const response = await fetch('/api/auth/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username, password })
                });
                
                const data = await response.json();
                if (response.ok) {
                    localStorage.setItem('token', data.token);
                    localStorage.setItem('user', JSON.stringify(data.user));
                    window.location.href = '/';
                } else {
                    alert(data.error || '登录失败');
                }
            } catch (error) {
                alert('登录失败: ' + error.message);
            }
        });
    </script>
</body>
</html>
{{end}}`

const indexHTML = `{{define "index"}}
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SecuArchive - 安全服务报告归档系统</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        :root {
            --primary-color: #1e3a5f;
            --accent-color: #00d4aa;
            --bg-color: #1a1a2e;
            --card-bg: #16213e;
            --text-color: #e8e8e8;
            --border-color: #2a2a4a;
        }
        * { box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-color);
            margin: 0;
            padding: 0;
            min-height: 100vh;
        }
        .sidebar {
            background: var(--card-bg);
            min-height: 100vh;
            border-right: 1px solid var(--border-color);
            padding: 20px 0;
            width: 240px;
            position: fixed;
            left: 0;
            top: 0;
        }
        .sidebar .nav-link {
            color: var(--text-color);
            padding: 12px 20px;
            margin: 5px 10px;
            border-radius: 8px;
            transition: all 0.3s;
            cursor: pointer;
        }
        .sidebar .nav-link:hover, .sidebar .nav-link.active {
            background: var(--accent-color);
            color: var(--bg-color);
        }
        .sidebar .nav-link i { margin-right: 10px; width: 20px; }
        .main-content {
            margin-left: 240px;
            padding: 20px;
            min-height: 100vh;
        }
        .top-bar {
            background: var(--card-bg);
            padding: 15px 20px;
            border-radius: 12px;
            margin-bottom: 20px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            border: 1px solid var(--border-color);
        }
        .card {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 12px;
            margin-bottom: 20px;
        }
        .card-header {
            background: transparent;
            border-bottom: 1px solid var(--border-color);
            padding: 15px 20px;
            font-weight: 600;
        }
        .table { color: var(--text-color); }
        .table thead th {
            border-bottom: 1px solid var(--border-color);
            color: var(--accent-color);
            font-weight: 600;
        }
        .table td { border-bottom: 1px solid var(--border-color); }
        .badge-category {
            background: var(--accent-color);
            color: var(--bg-color);
            padding: 5px 10px;
            border-radius: 15px;
            font-size: 12px;
        }
        .btn-action {
            padding: 5px 10px;
            border: none;
            border-radius: 5px;
            cursor: pointer;
            margin: 0 2px;
        }
        .btn-edit { background: #3498db; color: white; }
        .btn-delete { background: #e74c3c; color: white; }
        .btn-download { background: #2ecc71; color: white; }
        .search-box {
            background: var(--bg-color);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 10px 15px;
            color: var(--text-color);
            width: 300px;
        }
        .modal-content { background: var(--card-bg); border: 1px solid var(--border-color); }
        .form-label { color: var(--text-color); }
        .btn-close { filter: invert(1); }
        .nav-tabs .nav-link {
            color: var(--text-color);
            border: none;
            padding: 12px 20px;
            cursor: pointer;
        }
        .nav-tabs .nav-link.active {
            background: var(--accent-color);
            color: var(--bg-color);
            border-radius: 8px 8px 0 0;
        }
        .ai-chat-container {
            height: 400px;
            overflow-y: auto;
            background: var(--bg-color);
            border-radius: 8px;
            padding: 15px;
            margin-bottom: 15px;
        }
        .chat-message {
            margin-bottom: 15px;
            padding: 10px 15px;
            border-radius: 10px;
            max-width: 80%;
        }
        .chat-message.user {
            background: var(--accent-color);
            color: var(--bg-color);
            margin-left: auto;
        }
        .chat-message.ai {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
        }
        .file-drop-zone {
            border: 2px dashed var(--border-color);
            border-radius: 12px;
            padding: 40px;
            text-align: center;
            cursor: pointer;
            transition: all 0.3s;
        }
        .file-drop-zone:hover {
            border-color: var(--accent-color);
            background: rgba(0, 212, 170, 0.1);
        }
        .file-drop-zone i { font-size: 48px; color: var(--accent-color); margin-bottom: 15px; }
        .stats-card {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 12px;
            padding: 20px;
            text-align: center;
        }
        .stats-card h3 { font-size: 32px; color: var(--accent-color); margin: 10px 0; }
        .stats-card p { color: #888; margin: 0; }
        .loading { display: none; text-align: center; padding: 20px; }
        .loading i { font-size: 32px; color: var(--accent-color); animation: spin 1s linear infinite; }
        @keyframes spin { 100% { transform: rotate(360deg); } }
        ::-webkit-scrollbar { width: 8px; height: 8px; }
        ::-webkit-scrollbar-track { background: var(--bg-color); }
        ::-webkit-scrollbar-thumb { background: var(--border-color); border-radius: 4px; }
        .section { display: none; }
        .section.active { display: block; }
        .user-info { display: flex; align-items: center; gap: 10px; }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="text-center mb-4">
            <i class="fas fa-shield-halved fa-2x" style="color: var(--accent-color);"></i>
            <h5 class="mt-2">SecuArchive</h5>
        </div>
        <nav class="nav flex-column">
            <a class="nav-link active" data-section="dashboard"><i class="fas fa-home"></i>仪表盘</a>
            <a class="nav-link" data-section="projects"><i class="fas fa-folder"></i>项目管理</a>
            <a class="nav-link" data-section="reports"><i class="fas fa-file-alt"></i>报告管理</a>
            <a class="nav-link" data-section="templates"><i class="fas fa-file-import"></i>模板管理</a>
            <a class="nav-link" data-section="ai"><i class="fas fa-robot"></i>AI助手</a>
            <a class="nav-link" data-section="backup"><i class="fas fa-database"></i>备份管理</a>
            <a class="nav-link" data-section="settings"><i class="fas fa-cog"></i>系统设置</a>
        </nav>
    </div>
    
    <div class="main-content">
        <div class="top-bar">
            <h4 id="pageTitle">仪表盘</h4>
            <div class="user-info">
                <span id="userName"></span>
                <button class="btn btn-action btn-edit" onclick="logout()">
                    <i class="fas fa-sign-out-alt"></i>
                </button>
            </div>
        </div>
        
        <!-- Dashboard Section -->
        <div id="dashboard" class="section active">
            <div class="row mb-4">
                <div class="col-md-3">
                    <div class="stats-card">
                        <i class="fas fa-file-alt fa-2x" style="color: var(--accent-color);"></i>
                        <h3 id="reportCount">0</h3>
                        <p>报告总数</p>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="stats-card">
                        <i class="fas fa-folder fa-2x" style="color: var(--accent-color);"></i>
                        <h3 id="projectCount">0</h3>
                        <p>项目总数</p>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="stats-card">
                        <i class="fas fa-database fa-2x" style="color: var(--accent-color);"></i>
                        <h3 id="backupCount">0</h3>
                        <p>备份数量</p>
                    </div>
                </div>
                <div class="col-md-3">
                    <div class="stats-card">
                        <i class="fas fa-robot fa-2x" style="color: var(--accent-color);"></i>
                        <h3 id="aiStatus">未配置</h3>
                        <p>AI状态</p>
                    </div>
                </div>
            </div>
            <div class="card">
                <div class="card-header">最近报告</div>
                <div class="card-body">
                    <table class="table" id="recentReportsTable">
                        <thead><tr><th>报告名称</th><th>分类</th><th>项目</th><th>上传时间</th></tr></thead>
                        <tbody></tbody>
                    </table>
                </div>
            </div>
        </div>
        
        <!-- Reports Section -->
        <div id="reports" class="section">
            <div class="card">
                <div class="card-header d-flex justify-content-between align-items-center">
                    <span>报告列表</span>
                    <button class="btn btn-primary btn-sm" onclick="showImportModal()">
                        <i class="fas fa-plus"></i> 导入报告
                    </button>
                </div>
                <div class="card-body">
                    <div class="mb-3">
                        <input type="text" class="search-box" placeholder="搜索报告..." id="reportSearch" onkeyup="searchReports()">
                        <select class="form-select" style="width: 200px; display: inline-block; background: var(--card-bg); border: 1px solid var(--border-color); color: var(--text-color);" id="reportCategoryFilter" onchange="filterReports()">
                            <option value="">全部分类</option>
                        </select>
                    </div>
                    <table class="table" id="reportsTable">
                        <thead><tr><th>报告名称</th><th>分类</th><th>子分类</th><th>项目</th><th>大小</th><th>上传时间</th><th>操作</th></tr></thead>
                        <tbody></tbody>
                    </table>
                </div>
            </div>
        </div>
        
        <!-- Projects Section -->
        <div id="projects" class="section">
            <div class="card">
                <div class="card-header d-flex justify-content-between align-items-center">
                    <span>项目列表</span>
                    <button class="btn btn-primary btn-sm" onclick="showProjectModal()">
                        <i class="fas fa-plus"></i> 创建项目
                    </button>
                </div>
                <div class="card-body">
                    <table class="table" id="projectsTable">
                        <thead><tr><th>项目名称</th><th>分类</th><th>客户</th><th>状态</th><th>报告数</th><th>创建时间</th><th>操作</th></tr></thead>
                        <tbody></tbody>
                    </table>
                </div>
            </div>
        </div>
        
        <!-- AI Section -->
        <div id="ai" class="section">
            <div class="card">
                <div class="card-header d-flex justify-content-between align-items-center">
                    <span>AI 智能助手</span>
                    <button class="btn btn-sm btn-secondary" onclick="showAIConfigModal()">
                        <i class="fas fa-cog"></i> 设置
                    </button>
                </div>
                <div class="card-body">
                    <div class="ai-chat-container" id="chatContainer"></div>
                    <div class="input-group">
                        <input type="text" class="form-control" id="chatInput" placeholder="输入您的问题..." onkeypress="handleChatKeypress(event)">
                        <button class="btn btn-primary" onclick="sendChat()"><i class="fas fa-paper-plane"></i></button>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Templates Section -->
        <div id="templates" class="section">
            <div class="card">
                <div class="card-header d-flex justify-content-between align-items-center">
                    <span>模板列表</span>
                    <button class="btn btn-primary btn-sm" onclick="showImportTemplateModal()">
                        <i class="fas fa-plus"></i> 导入模板
                    </button>
                </div>
                <div class="card-body">
                    <table class="table" id="templatesTable">
                        <thead><tr><th>模板名称</th><th>分类</th><th>文件名</th><th>大小</th><th>上传时间</th><th>操作</th></tr></thead>
                        <tbody></tbody>
                    </table>
                </div>
            </div>
        </div>
        
        <!-- Backup Section -->
        <div id="backup" class="section">
            <div class="card">
                <div class="card-header d-flex justify-content-between align-items-center">
                    <span>备份管理</span>
                    <div>
                        <button class="btn btn-primary btn-sm" onclick="createBackup()">
                            <i class="fas fa-plus"></i> 创建备份
                        </button>
                        <button class="btn btn-secondary btn-sm" onclick="document.getElementById('importBackupInput').click()">
                            <i class="fas fa-upload"></i> 导入备份
                        </button>
                        <input type="file" id="importBackupInput" style="display: none;" onchange="importBackup(this)">
                    </div>
                </div>
                <div class="card-body">
                    <table class="table" id="backupsTable">
                        <thead><tr><th>文件名</th><th>大小</th><th>类型</th><th>创建时间</th><th>操作</th></tr></thead>
                        <tbody></tbody>
                    </table>
                </div>
            </div>
        </div>
        
        <!-- Settings Section -->
        <div id="settings" class="section">
            <div class="card">
                <div class="card-header">用户管理</div>
                <div class="card-body">
                    <div class="mb-3">
                        <label class="form-label">修改密码</label>
                        <input type="password" class="form-control" id="newPassword" placeholder="新密码">
                    </div>
                    <button type="button" class="btn btn-primary" onclick="updatePassword()">修改密码</button>
                </div>
            </div>
        </div>
    </div>
    
    <!-- Import Modal -->
    <div class="modal fade" id="importModal" tabindex="-1">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">导入报告</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <div class="file-drop-zone" onclick="document.getElementById('reportFile').click()">
                        <i class="fas fa-cloud-upload-alt"></i>
                        <p>点击或拖拽文件到此处上传</p>
                        <input type="file" id="reportFile" style="display: none;" onchange="handleFileSelect(this)">
                    </div>
                    <div id="selectedFileName" class="mt-2 text-center" style="color: var(--accent-color);"></div>
                    <div class="mb-3 mt-3">
                        <label class="form-label">报告名称</label>
                        <input type="text" class="form-control" id="reportName">
                    </div>
                    <div class="mb-3">
                        <label class="form-label">所属项目</label>
                        <select class="form-select" id="reportProject" style="background: var(--bg-color); border: 1px solid var(--border-color); color: var(--text-color);"></select>
                    </div>
                    <div class="mb-3">
                        <label class="form-label">分类</label>
                        <select class="form-select" id="reportCategory" style="background: var(--bg-color); border: 1px solid var(--border-color); color: var(--text-color);" onchange="updateSubCategories()">
                            <option value="">请选择分类</option>
                        </select>
                    </div>
                    <div class="mb-3">
                        <label class="form-label">子分类</label>
                        <select class="form-select" id="reportSubCategory" style="background: var(--bg-color); border: 1px solid var(--border-color); color: var(--text-color);"></select>
                    </div>
                    <div class="mb-3">
                        <label class="form-label">标签 (用逗号分隔)</label>
                        <input type="text" class="form-control" id="reportTags" placeholder="标签1, 标签2">
                    </div>
                    <div class="mb-3">
                        <label class="form-label">描述</label>
                        <textarea class="form-control" id="reportDescription" rows="3"></textarea>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                    <button type="button" class="btn btn-primary" onclick="importReport()">导入</button>
                </div>
            </div>
        </div>
    </div>
    
    <!-- Project Modal -->
    <div class="modal fade" id="projectModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title" id="projectModalTitle">创建项目</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <input type="hidden" id="projectId">
                    <div class="row">
                        <div class="col-md-6 mb-3">
                            <label class="form-label">项目名称 *</label>
                            <input type="text" class="form-control" id="projectName">
                        </div>
                        <div class="col-md-6 mb-3">
                            <label class="form-label">分类</label>
                            <select class="form-select" id="projectCategory" style="background: var(--bg-color); border: 1px solid var(--border-color); color: var(--text-color);">
                                <option value="渗透测试">渗透测试</option>
                                <option value="代码审计">代码审计</option>
                                <option value="安全评估">安全评估</option>
                                <option value="应急响应">应急响应</option>
                                <option value="风险评估">风险评估</option>
                                <option value="合规审计">合规审计</option>
                                <option value="其他">其他</option>
                            </select>
                        </div>
                    </div>
                    <div class="row">
                        <div class="col-md-6 mb-3">
                            <label class="form-label">客户名称</label>
                            <input type="text" class="form-control" id="projectClient">
                        </div>
                        <div class="col-md-6 mb-3">
                            <label class="form-label">合同所属</label>
                            <input type="text" class="form-control" id="projectContract" placeholder="合同所属公司/部门">
                        </div>
                    </div>
                    <div class="row">
                        <div class="col-md-6 mb-3">
                            <label class="form-label">合同编号</label>
                            <input type="text" class="form-control" id="projectContractNo" placeholder="合同编号">
                        </div>
                        <div class="col-md-6 mb-3">
                            <label class="form-label">导入报告压缩包 (可选)</label>
                            <input type="file" class="form-control" id="projectZipFile" accept=".zip">
                            <small class="text-muted">支持 .zip 格式，会自动解压并智能分类</small>
                        </div>
                    </div>
                    <div class="mb-3">
                        <label class="form-label">描述</label>
                        <textarea class="form-control" id="projectDescription" rows="3"></textarea>
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                    <button type="button" class="btn btn-primary" onclick="saveProject()">保存</button>
                </div>
            </div>
        </div>
    </div>
    
    <!-- Template Import Modal -->
    <div class="modal fade" id="templateModal" tabindex="-1">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">导入模板</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <div class="mb-3">
                        <label class="form-label">模板名称</label>
                        <input type="text" class="form-control" id="templateName">
                    </div>
                    <div class="mb-3">
                        <label class="form-label">分类</label>
                        <select class="form-select" id="templateCategory" style="background: var(--bg-color); border: 1px solid var(--border-color); color: var(--text-color);">
                            <option value="渗透测试">渗透测试</option>
                            <option value="代码审计">代码审计</option>
                            <option value="基线核查">基线核查</option>
                            <option value="漏洞扫描">漏洞扫描</option>
                            <option value="安全评估">安全评估</option>
                            <option value="应急响应">应急响应</option>
                            <option value="风险评估">风险评估</option>
                            <option value="合规审计">合规审计</option>
                            <option value="其他">其他</option>
                        </select>
                    </div>
                    <div class="mb-3">
                        <label class="form-label">描述</label>
                        <textarea class="form-control" id="templateDescription" rows="2"></textarea>
                    </div>
                    <div class="mb-3">
                        <label class="form-label">模板文件</label>
                        <input type="file" class="form-control" id="templateFile">
                    </div>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">取消</button>
                    <button type="button" class="btn btn-primary" onclick="importTemplate()">导入</button>
                </div>
            </div>
        </div>
    </div>
    
    <!-- AI Config Modal -->
    <div class="modal fade" id="aiConfigModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">AI 模型配置</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <form id="aiConfigForm">
                        <div class="row">
                            <div class="col-md-6 mb-3">
                                <label class="form-label">AI 提供商</label>
                                <select class="form-select" id="aiProvider" style="background: var(--bg-color); border: 1px solid var(--border-color); color: var(--text-color);" onchange="onProviderChange()">
                                    <option value="">加载中...</option>
                                </select>
                            </div>
                            <div class="col-md-6 mb-3">
                                <label class="form-label">模型</label>
                                <select class="form-select" id="aiModel" style="background: var(--bg-color); border: 1px solid var(--border-color); color: var(--text-color);">
                                    <option value="">请先选择提供商</option>
                                </select>
                            </div>
                        </div>
                        <div class="mb-3">
                            <label class="form-label">
                                <input type="checkbox" id="aiUseCustomModel" onchange="toggleCustomModel()"> 使用自定义模型
                            </label>
                        </div>
                        <div class="mb-3" id="customModelSection" style="display: none;">
                            <label class="form-label">自定义模型名称</label>
                            <input type="text" class="form-control" id="aiCustomModel" placeholder="请输入自定义模型名称">
                        </div>
                        <div class="mb-3">
                            <label class="form-label">API Key</label>
                            <input type="password" class="form-control" id="aiApiKey" placeholder="请输入 API Key">
                        </div>
                        <div class="mb-3">
                            <label class="form-label">API 端点 (可选)</label>
                            <input type="text" class="form-control" id="aiEndpoint" placeholder="留空使用默认端点">
                            <small class="text-muted">自定义API代理地址，如使用代理软件时填写</small>
                        </div>
                        <div class="mb-3">
                            <label class="form-check-label">
                                <input type="checkbox" class="form-check-input" id="aiEnabled"> 启用 AI
                            </label>
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" onclick="testAIConnection()">测试连接</button>
                    <button type="button" class="btn btn-primary" onclick="saveAIConfig()">保存配置</button>
                </div>
            </div>
        </div>
    </div>
    
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script src="/static/js/app.js"></script>
</body>
</html>
{{end}}`
