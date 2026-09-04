<template>
  <div class="login-page">
    <!-- Left: Brand Panel -->
    <div class="brand-panel">
      <div class="brand-bg-pattern">
        <div class="circle circle-1"></div>
        <div class="circle circle-2"></div>
        <div class="circle circle-3"></div>
      </div>
      <div class="brand-content">
        <div class="brand-logo">
          <div class="logo-icon">B</div>
          <span class="logo-name">ByteCode</span>
        </div>
        <h1 class="brand-headline">AI-Native<br/>项目管理平台</h1>
        <p class="brand-desc">
          集成产品管理、任务追踪、测试管理与 AI 智能体，<br/>
          开箱即用、零外部依赖。
        </p>
        <div class="brand-features">
          <div class="feature-item">
            <div class="feature-dot"></div>
            <span>智能任务分配与追踪</span>
          </div>
          <div class="feature-item">
            <div class="feature-dot"></div>
            <span>可视化数据库模型管理</span>
          </div>
          <div class="feature-item">
            <div class="feature-dot"></div>
            <span>全流程测试管理</span>
          </div>
        </div>
      </div>
      <div class="brand-footer">
        <span>&copy; {{ new Date().getFullYear() }} ByteCode</span>
      </div>
    </div>

    <!-- Right: Login Form -->
    <div class="form-panel">
      <div class="form-container">
        <div class="form-header">
          <h2 class="form-title">欢迎回来</h2>
          <p class="form-subtitle">登录您的账号以继续</p>
        </div>

        <n-form
          ref="formRef"
          label-placement="top"
          size="large"
          :model="formInline"
          :rules="rules"
          class="login-form"
        >
          <n-form-item label="用户名" path="username">
            <n-input
              v-model:value="formInline.username"
              placeholder="请输入用户名"
              class="login-input"
            >
              <template #prefix>
                <n-icon size="18" color="#94a3b8">
                  <PersonOutline />
                </n-icon>
              </template>
            </n-input>
          </n-form-item>

          <n-form-item label="密码" path="password">
            <n-input
              v-model:value="formInline.password"
              type="password"
              showPasswordOn="click"
              placeholder="请输入密码"
              class="login-input"
            >
              <template #prefix>
                <n-icon size="18" color="#94a3b8">
                  <LockClosedOutline />
                </n-icon>
              </template>
            </n-input>
          </n-form-item>

          <div class="form-options">
            <n-checkbox v-model:checked="autoLogin">记住登录</n-checkbox>
          </div>

          <n-button
            type="primary"
            @click="handleSubmit"
            size="large"
            :loading="loading"
            block
            class="login-button"
          >
            登录
          </n-button>
        </n-form>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { reactive, ref, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useUserStore } from '@/store/modules/user';
  import { forcePwdVisible } from '@/store/modules/forcePassword';
  import { useMessage } from 'naive-ui';
  import { ResultEnum } from '@/enums/httpEnum';
  import { PersonOutline, LockClosedOutline } from '@vicons/ionicons5';
  import { PageEnum } from '@/enums/pageEnum';

  onMounted(() => {
    setTimeout(() => {
      const usernameInput = document.querySelector('input[placeholder="请输入用户名"]');
      if (usernameInput) {
        (usernameInput as HTMLElement).focus();
      }
    }, 300);
  });

  interface FormState {
    username: string;
    password: string;
  }

  const formRef = ref();
  const message = useMessage();
  const loading = ref(false);
  const autoLogin = ref(true);
  const LOGIN_NAME = PageEnum.BASE_LOGIN_NAME;

  const formInline = reactive({
    username: '',
    password: '',
    isCaptcha: true,
  });

  const rules = {
    username: { required: true, message: '请输入用户名', trigger: 'blur' },
    password: { required: true, message: '请输入密码', trigger: 'blur' },
  };

  const userStore = useUserStore();
  const router = useRouter();
  const route = useRoute();

  const handleSubmit = (e) => {
    e.preventDefault();
    formRef.value.validate(async (errors) => {
      if (!errors) {
        const { username, password } = formInline;
        message.loading('登录中...');
        loading.value = true;

        const params: FormState = { username, password };

        try {
          const { code, message: msg, result } = await userStore.login(params);
          message.destroyAll();
          if (code == ResultEnum.SUCCESS) {
            // 仍在使用初始默认密码的账号，进入系统后由全局弹窗接管改密
            if (result?.mustChangePassword) {
              forcePwdVisible.value = true;
            }
            const toPath = decodeURIComponent((route.query?.redirect || '/') as string);
            message.success('登录成功，即将进入系统');
            if (route.name === LOGIN_NAME) {
              router.replace('/');
            } else router.replace(toPath);
          } else {
            message.info(msg || '登录失败');
          }
        } finally {
          loading.value = false;
        }
      } else {
        message.error('请填写完整登录信息');
      }
    });
  };
</script>

<style lang="less" scoped>
  .login-page {
    display: flex;
    height: 100vh;
    overflow: hidden;
    background: var(--canvas, #f8fafc);
  }

  /* ─── Left: Brand Panel ─── */
  .brand-panel {
    position: relative;
    flex: 0 0 45%;
    background: linear-gradient(135deg, #1e40af 0%, #2563eb 50%, #3b82f6 100%);
    display: flex;
    flex-direction: column;
    justify-content: center;
    padding: 60px 56px;
    overflow: hidden;

    .brand-bg-pattern {
      position: absolute;
      inset: 0;
      pointer-events: none;
    }

    .circle {
      position: absolute;
      border-radius: 50%;
      opacity: 0.08;
      background: var(--panel-bg, #fff);

      &-1 {
        width: 400px;
        height: 400px;
        top: -120px;
        left: -100px;
      }
      &-2 {
        width: 300px;
        height: 300px;
        bottom: -80px;
        right: -60px;
      }
      &-3 {
        width: 150px;
        height: 150px;
        top: 50%;
        right: 15%;
        opacity: 0.05;
      }
    }

    .brand-content {
      position: relative;
      z-index: 1;
    }

    .brand-logo {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 48px;

      .logo-icon {
        width: 44px;
        height: 44px;
        border-radius: 12px;
        background: rgba(255, 255, 255, 0.2);
        backdrop-filter: blur(10px);
        border: 1px solid rgba(255, 255, 255, 0.15);
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 22px;
        font-weight: 700;
        color: #fff;
      }

      .logo-name {
        font-size: 22px;
        font-weight: 600;
        color: #fff;
        letter-spacing: 0.5px;
      }
    }

    .brand-headline {
      font-size: 38px;
      font-weight: 700;
      color: #fff;
      line-height: 1.3;
      margin: 0 0 20px;
    }

    .brand-desc {
      font-size: 15px;
      color: rgba(255, 255, 255, 0.75);
      line-height: 1.7;
      margin: 0 0 40px;
    }

    .brand-features {
      display: flex;
      flex-direction: column;
      gap: 16px;

      .feature-item {
        display: flex;
        align-items: center;
        gap: 12px;
        color: rgba(255, 255, 255, 0.85);
        font-size: 14px;

        .feature-dot {
          width: 6px;
          height: 6px;
          border-radius: 50%;
          background: rgba(255, 255, 255, 0.6);
          flex-shrink: 0;
        }
      }
    }

    .brand-footer {
      position: absolute;
      bottom: 32px;
      left: 56px;
      right: 56px;
      z-index: 1;

      span {
        font-size: 13px;
        color: rgba(255, 255, 255, 0.4);
      }
    }
  }

  /* ─── Right: Form Panel ─── */
  .form-panel {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 40px;
    // 跟随明暗主题（暗色下白色输入文字需要深色底才可见）
    background: var(--canvas, #f8fafc);
  }

  .form-container {
    width: 100%;
    max-width: 400px;
  }

  .form-header {
    margin-bottom: 36px;

    .form-title {
      font-size: 26px;
      font-weight: 700;
      color: var(--text-1, #24292f);
      margin: 0 0 8px;
    }

    .form-subtitle {
      font-size: 15px;
      color: var(--text-2, #57606a);
      margin: 0;
    }
  }

  .login-form {
    :deep(.n-form-item-label) {
      font-size: 14px;
      font-weight: 500;
      color: var(--text-2, #57606a);
      padding-bottom: 6px;
    }

    :deep(.n-form-item) {
      margin-bottom: 20px;
    }

    :deep(.n-form-item-feedback-wrapper) {
      min-height: 18px;
    }

    :deep(.n-input) {
      border-radius: 8px;
    }

    :deep(.n-input .n-input__border),
    :deep(.n-input .n-input__state-border) {
      border-radius: 8px;
    }
  }

  .login-input {
    // 输入元素与 placeholder 覆盖层须同 padding，否则空态光标(5px)与提示文字(0px)错位
    :deep(.n-input__input-el),
    :deep(.n-input__placeholder) {
      padding-left: 5px;
    }
  }

  .form-options {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    font-size: 14px;
  }

  .login-button {
    height: 44px;
    font-size: 16px;
    font-weight: 500;
    border-radius: 8px;
    transition: all 0.2s ease;
  }

  /* ─── Responsive ─── */
  @media (max-width: 1024px) {
    .brand-panel {
      flex: 0 0 40%;
      padding: 40px;

      .brand-headline {
        font-size: 30px;
      }

      .brand-footer {
        left: 40px;
        right: 40px;
      }
    }
  }

  @media (max-width: 768px) {
    .login-page {
      flex-direction: column;
    }

    .brand-panel {
      flex: none;
      padding: 32px 24px;
      min-height: auto;

      .brand-headline {
        font-size: 24px;
        margin-bottom: 12px;
      }

      .brand-desc {
        display: none;
      }

      .brand-features {
        display: none;
      }

      .brand-logo {
        margin-bottom: 16px;
      }

      .brand-footer {
        display: none;
      }
    }

    .form-panel {
      flex: 1;
      padding: 32px 24px;
      align-items: flex-start;
      padding-top: 24px;
    }
  }

  @media (max-width: 480px) {
    .brand-panel {
      padding: 24px 20px;

      .brand-headline {
        font-size: 20px;
      }
    }

    .form-panel {
      padding: 24px 20px;
    }
  }
</style>
