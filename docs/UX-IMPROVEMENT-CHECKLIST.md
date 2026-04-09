# My PaaS — UI/UX Improvement Checklist

> Đánh giá từ góc nhìn khách hàng cũ Railway. Ưu tiên theo mức độ ảnh hưởng.

---

## CRITICAL — Phải sửa

- [x] **C1: Responsive Sidebar (Mobile/Tablet)** ✅
  - Sidebar cố định w-56 (224px) chiếm >50% viewport mobile
  - Done: hamburger menu toggle, sidebar collapse, overlay trên mobile
  - Files: `web/src/components/Layout.tsx`

- [x] **C2: Deploy Feedback — Loading + Streaming** ✅
  - Click Deploy → không có visual progress, chỉ "Deploying..." text
  - Done: auto-expand deployment row, Loader2 spinner, polling updates
  - Files: `web/src/pages/ProjectDetail.tsx`

- [x] **C3: Unique Project Names** ✅
  - Dashboard hiển thị trùng "demo-node" x2, không phân biệt được
  - Done: backend 409 duplicate check + frontend inline error
  - Files: `server/handler/project.go`, `web/src/pages/NewProject.tsx`

- [x] **C4: Type-to-Confirm Delete** ✅
  - Delete Project chỉ dùng `confirm()` browser native
  - Done: custom DeleteModal yêu cầu type project name
  - Files: `web/src/pages/ProjectDetail.tsx`

---

## HIGH — Ảnh hưởng trải nghiệm đáng kể

- [x] **H1: Search & Filter Projects** ✅
  - 8+ projects đã khó tìm, scale lên 50+ không dùng được
  - Done: search bar + filter by status (Healthy/Failed/Building)
  - Files: `web/src/pages/Dashboard.tsx`

- [x] **H5: Inline Form Validation** ✅
  - Error hiện cuối form, cách xa field lỗi. Chỉ hiện 1 lỗi/lần.
  - Done: inline error ngay dưới mỗi field, red border, clear on focus
  - Files: `web/src/pages/NewProject.tsx`

- [x] **H6: Project URL / Domain Display** ✅
  - Project detail không hiện URL app đang chạy ở đâu
  - Done: hiện domain links hoặc internal service name badge
  - Files: `web/src/pages/ProjectDetail.tsx`

---

## MEDIUM — Cải thiện chung

- [x] **M1: Dashboard Card Hover Effect** ✅
  - Project cards hover quá nhẹ, thiếu visual feedback
  - Done: border-accent glow + shadow + transition-all 200ms
  - Files: `web/src/pages/Dashboard.tsx`

- [x] **M9: Build Logs Color Coding** ✅
  - Logs monochrome, khó phân biệt error/success/info
  - Done: error/warn lines get colored bg + left border, success/fail badge in header
  - Files: `web/src/components/LogViewer.tsx`

---

## Implementation Order

1. **Layout.tsx** — Responsive sidebar (C1)
2. **ProjectDetail.tsx** — Deploy streaming (C2) + Delete modal (C4) + URL display (H6)
3. **Dashboard.tsx** — Search/filter (H1) + Hover effects (M1)
4. **NewProject.tsx** — Inline validation (H5)
5. **LogViewer.tsx** — Color coding (M9)
6. **Backend** — Unique project name constraint (C3)
7. **Build & Deploy** — Verify all changes on VM
