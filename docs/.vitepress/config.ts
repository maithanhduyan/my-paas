import { defineConfig } from 'vitepress'

export default defineConfig({
  lang: 'vi-VN',
  title: 'My PaaS',
  description: 'Nền tảng triển khai ứng dụng tự quản — Platform as a Service mã nguồn mở',

  head: [
    ['link', { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' }],
  ],

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: 'Tài liệu', link: '/guide/introduction' },
      { text: 'Kiến trúc', link: '/architecture/overview' },
      { text: 'API', link: '/api/endpoints' },
      { text: 'Roadmap', link: '/roadmap' },
    ],

    sidebar: [
      {
        text: 'Bắt đầu',
        items: [
          { text: 'Giới thiệu', link: '/guide/introduction' },
          { text: 'Tại sao My PaaS?', link: '/guide/why' },
          { text: 'Cài đặt & Triển khai', link: '/guide/deployment' },
          { text: 'Cấu hình', link: '/guide/configuration' },
        ],
      },
      {
        text: 'Kiến trúc',
        items: [
          { text: 'Tổng quan hệ thống', link: '/architecture/overview' },
          { text: 'Lựa chọn công nghệ', link: '/architecture/tech-choices' },
          { text: 'Pipeline triển khai', link: '/architecture/deploy-pipeline' },
          { text: 'Enterprise', link: '/architecture/enterprise' },
        ],
      },
      {
        text: 'Tính năng',
        items: [
          { text: 'Quản lý dự án', link: '/features/projects' },
          { text: 'Dịch vụ hỗ trợ', link: '/features/services' },
          { text: 'Docker Registry', link: '/features/registry' },
          { text: 'CLI Tool', link: '/features/cli' },
          { text: 'VS Code Extension', link: '/features/vscode-extension' },
        ],
      },
      {
        text: 'Tiêu chuẩn',
        items: [
          { text: 'Tiêu chuẩn chất lượng', link: '/standards/quality' },
          { text: 'Bảo mật', link: '/standards/security' },
        ],
      },
      {
        text: 'API Reference',
        items: [
          { text: 'Tổng quan API', link: '/api/endpoints' },
        ],
      },
      {
        text: 'Kế hoạch',
        items: [
          { text: 'Roadmap', link: '/roadmap' },
        ],
      },
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/my-paas/my-paas' },
    ],

    outline: {
      level: [2, 3],
      label: 'Mục lục',
    },

    footer: {
      message: 'My PaaS — Open Source Platform as a Service',
      copyright: '© 2026 My PaaS Project',
    },

    search: {
      provider: 'local',
      options: {
        translations: {
          button: { buttonText: 'Tìm kiếm' },
          modal: {
            noResultsText: 'Không tìm thấy kết quả',
            resetButtonTitle: 'Xoá tìm kiếm',
          },
        },
      },
    },

    docFooter: {
      prev: 'Trang trước',
      next: 'Trang sau',
    },
  },

  markdown: {
    math: true,
  },
})
