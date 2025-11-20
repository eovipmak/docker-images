# V-Insight Custom GitHub Copilot Agents

Bộ custom agents chuyên biệt cho dự án V-Insight - Multi-tenant Monitoring SaaS Platform.

## Tổng quan

Repository này chứa 6 custom agents được thiết kế đặc biệt cho kiến trúc và công nghệ của V-Insight:

1. **v-insight-architect** - Kiến trúc sư hệ thống tổng quan
2. **backend-specialist** - Chuyên gia Go backend (API + Worker)
3. **frontend-specialist** - Chuyên gia SvelteKit frontend
4. **database-specialist** - Chuyên gia PostgreSQL và migrations
5. **docker-devops-specialist** - Chuyên gia Docker và DevOps
6. **testing-specialist** - Chuyên gia testing (unit, integration, E2E)

## Cài đặt

### Cách 1: Cài đặt vào Repository (Khuyến nghị)

1. Tạo thư mục `.github/agents` trong repository của bạn:
```bash
mkdir -p .github/agents
```

2. Copy các file agent vào thư mục này:
```bash
cp *.agent.md .github/agents/
```

3. Commit và push lên GitHub:
```bash
git add .github/agents/
git commit -m "Add custom Copilot agents"
git push
```

4. Truy cập [GitHub Copilot Agents](https://github.com/copilot/agents) và chọn repository của bạn để thấy các agents.

### Cách 2: Cài đặt trong VS Code (Local)

1. Mở GitHub Copilot Chat trong VS Code
2. Click vào dropdown agents → **Configure Custom Agents...**
3. Click **Create new custom agent**
4. Chọn **Workspace** để tạo agent cho workspace hiện tại
5. Copy nội dung từ các file `.agent.md` vào

### Cách 3: Cài đặt trong JetBrains IDEs

1. Mở GitHub Copilot Chat
2. Click vào dropdown agents → **Configure Agents...**
3. Trong settings, chọn **Workspace**
4. Copy nội dung từ các file `.agent.md` vào

## Hướng dẫn sử dụng

### 1. V-Insight Architect

**Khi nào sử dụng:**
- Cần tư vấn về kiến trúc tổng thể
- Quyết định công nghệ và design patterns
- Thiết kế tính năng mới liên quan nhiều services
- Đánh giá trade-offs và giải pháp thay thế

**Ví dụ prompts:**
```
@v-insight-architect Tôi muốn thêm tính năng webhook notifications. Nên thiết kế như thế nào?

@v-insight-architect Làm thế nào để scale hệ thống từ 10K lên 100K users?

@v-insight-architect Đánh giá việc tách worker service thành microservice riêng
```

### 2. Backend Specialist

**Khi nào sử dụng:**
- Phát triển API endpoints (Gin framework)
- Xây dựng business logic
- Tạo background jobs cho Worker
- Debug backend issues

**Ví dụ prompts:**
```
@backend-specialist Tạo API endpoint để tạo monitor mới với validation

@backend-specialist Implement background job để check monitors mỗi 5 phút

@backend-specialist Làm sao để handle graceful shutdown trong Go service?

@backend-specialist Review code authentication middleware này
```

### 3. Frontend Specialist

**Khi nào sử dụng:**
- Phát triển UI components (Svelte)
- Cấu hình routing và pages
- Xử lý API integration qua proxy
- Debug frontend issues

**Ví dụ prompts:**
```
@frontend-specialist Tạo dashboard component để hiển thị monitor status

@frontend-specialist Setup server-side proxy để call backend API

@frontend-specialist Implement form validation cho login page

@frontend-specialist Optimize loading state và skeleton screens
```

### 4. Database Specialist

**Khi nào sử dụng:**
- Thiết kế database schema
- Viết migrations
- Tối ưu hóa queries
- Giải quyết vấn đề performance

**Ví dụ prompts:**
```
@database-specialist Tạo migration cho bảng monitors với multi-tenant support

@database-specialist Thiết kế schema cho storing monitoring check results

@database-specialist Tối ưu query này, nó chạy quá chậm: SELECT ...

@database-specialist Review index strategy cho bảng monitor_checks
```

### 5. Docker DevOps Specialist

**Khi nào sử dụng:**
- Cấu hình Docker và docker-compose
- Setup CI/CD pipeline
- Deploy và troubleshoot
- Monitoring và logging

**Ví dụ prompts:**
```
@docker-devops-specialist Setup health check cho backend service

@docker-devops-specialist Tối ưu Dockerfile để giảm image size

@docker-devops-specialist Tạo GitHub Actions workflow cho auto-deploy

@docker-devops-specialist Container backend không start được, giúp debug
```

### 6. Testing Specialist

**Khi nào sử dụng:**
- Viết unit tests
- Tạo integration tests
- Setup E2E testing
- Đảm bảo test coverage

**Ví dụ prompts:**
```
@testing-specialist Viết unit tests cho UserService.CreateUser

@testing-specialist Tạo integration test cho user registration flow

@testing-specialist Setup Playwright test cho login functionality

@testing-specialist Review test coverage và suggest improvements
```

## Best Practices

### 1. Chọn Agent phù hợp
- Sử dụng **architect** cho quyết định high-level
- Sử dụng **specialists** cho implementation cụ thể
- Có thể kết hợp nhiều agents cho các tác vụ phức tạp

### 2. Viết Prompts rõ ràng
```
✅ Tốt: "Tạo API endpoint POST /api/v1/monitors với validation cho URL và interval"

❌ Không tốt: "Làm API monitors"
```

### 3. Cung cấp Context
```
✅ Tốt: "Backend đang dùng Gin framework. Tôi muốn thêm middleware để log requests"

❌ Không tốt: "Thêm logging"
```

### 4. Yêu cầu Review
```
@backend-specialist Review code này và suggest improvements:
[paste code]
```

### 5. Học từ Output
- Agents được train với best practices
- Học patterns và conventions từ code suggestions
- Apply những patterns này vào code khác

## Cấu trúc Agent Files

Mỗi agent file có cấu trúc:

```markdown
---
name: agent-name
description: Brief description
tools: ["read", "edit", "search", "run"]
---

# Agent content with instructions and examples
```

### YAML Frontmatter Properties

- **name**: Tên agent (hiển thị trong dropdown)
- **description**: Mô tả ngắn gọn chức năng
- **tools**: Các tools agent có thể sử dụng
  - `read`: Đọc files
  - `edit`: Chỉnh sửa files
  - `search`: Tìm kiếm code
  - `run`: Chạy commands (cho testing)

## Tích hợp với Workflow

### Development Workflow

```mermaid
graph LR
    A[Plan Feature] --> B[@architect: Design]
    B --> C[@database: Schema]
    C --> D[@backend: API]
    D --> E[@frontend: UI]
    E --> F[@testing: Tests]
    F --> G[@docker: Deploy]
```

### Code Review Workflow

```
1. Write code manually
2. @specialist: Review this code
3. Apply suggestions
4. @testing-specialist: Add tests
5. Commit
```

## Troubleshooting

### Agent không hiển thị
- Đảm bảo file có extension `.agent.md`
- Check YAML frontmatter format đúng
- Refresh browser hoặc restart IDE

### Agent không hiểu context
- Cung cấp thêm context trong prompt
- Reference specific files hoặc code
- Describe current state và desired outcome

### Suggestions không phù hợp
- Kiểm tra đã chọn đúng agent
- Cung cấp requirements rõ ràng hơn
- Ask agent to explain reasoning

## Customization

### Modify Existing Agents

1. Edit file `.agent.md` tương ứng
2. Update instructions hoặc examples
3. Save và commit
4. Agents sẽ update automatically

### Create New Agents

1. Copy template từ agent hiện có
2. Customize name, description, tools
3. Add specific instructions
4. Test với example prompts
5. Add to `.github/agents/`

### Agent Templates

```markdown
---
name: my-custom-agent
description: What this agent does
tools: ["read", "edit", "search"]
---

You are a specialist for [specific domain].

## Responsibilities
- Task 1
- Task 2

## Guidelines
- Guideline 1
- Guideline 2

## Examples
[Provide examples]
```

## Resources

### Documentation
- [GitHub Copilot Custom Agents](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/coding-agent/create-custom-agents)
- [Custom Agents Configuration](https://docs.github.com/en/copilot/reference/custom-agents-configuration)
- [VS Code Custom Agents](https://code.visualstudio.com/docs/copilot/customization/custom-agents)

### V-Insight Project
- [Repository](https://github.com/eovipmak/v-insight)
- Architecture: Go + SvelteKit + PostgreSQL
- Docker Compose setup
- Multi-tenant SaaS platform

### Community
- [Awesome Copilot Agents](https://github.com/github/awesome-copilot/tree/main/agents)
- [Customization Library](https://docs.github.com/en/copilot/tutorials/customization-library/custom-agents)

## Version History

### v1.0.0 (Initial Release)
- 6 specialized agents
- Complete V-Insight coverage
- Vietnamese documentation
- Best practices and examples

## Contributing

Nếu bạn có suggestions để improve agents:

1. Test thoroughly với real use cases
2. Document changes clearly
3. Update examples if needed
4. Submit PR hoặc create issue

## License

Các agents này được tạo cho dự án V-Insight. Bạn có thể customize và sử dụng cho dự án của mình.

---

**Happy Coding with GitHub Copilot! 🚀**
