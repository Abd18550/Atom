import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../views/LoginView.vue'
import WelcomeView from '../views/WelcomeView.vue'
import AdminView from '../views/AdminView.vue'
import MentorView from '../views/MentorView.vue'
import ProfileView from '../views/ProfileView.vue'
import GroupsView from '../views/GroupsView.vue'
import GroupDetailsView from '../views/GroupDetailsView.vue'
import ExerciseListView from '../views/ExerciseListView.vue'
import ExerciseView from '../views/ExerciseView.vue'
import AdminExercisesView from '../views/AdminExercisesView.vue'
import LearningPathView from '../views/LearningPathView.vue'
import AdminStagesView from '../views/AdminStagesView.vue'
import AdminStageQuestionsView from '../views/AdminStageQuestionsView.vue'
import StageQuestionSandboxView from '../views/StageQuestionSandboxView.vue'
import StudentHomeView from '../views/StudentHomeView.vue'

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: '/',
            redirect: '/login'
        },
        {
            path: '/exercises',
            name: 'exercises',
            component: ExerciseListView,
            meta: { requiresAuth: true }
        },
        {
            path: '/learning-path',
            name: 'learning-path',
            component: LearningPathView,
            meta: { requiresAuth: true }
        },
        {
            path: '/exercises/:id',
            name: 'exercise',
            component: ExerciseView,
            meta: { requiresAuth: true }
        },
        {
            path: '/stage-questions/:id',
            name: 'stage-question-sandbox',
            component: StageQuestionSandboxView,
            meta: { requiresAuth: true }
        },
        {
            path: '/stage/:id',
            name: 'stage-orbit',
            component: () => import('../views/StageOrbitView.vue'),
            meta: { requiresAuth: true }
        },
        {
            path: '/login',
            name: 'login',
            component: LoginView
        },
        {
            path: '/welcome',
            name: 'welcome',
            component: WelcomeView,
            meta: { requiresAuth: true }
        },
        {
            path: '/student-home',
            name: 'student-home',
            component: StudentHomeView,
            meta: { requiresAuth: true }
        },
        {
            path: '/profile',
            name: 'profile',
            component: ProfileView,
            meta: { requiresAuth: true }
        },
        {
            path: '/admin',
            name: 'admin',
            component: AdminView,
            meta: { requiresAuth: true, roles: ['Admin', 'Supervisor'] }
        },
        {
            path: '/admin/exercises',
            name: 'admin-exercises',
            component: AdminExercisesView,
            meta: { requiresAuth: true, roles: ['Admin', 'Supervisor'] }
        },
        {
            path: '/admin/stages',
            name: 'admin-stages',
            component: AdminStagesView,
            meta: { requiresAuth: true, roles: ['Admin', 'Supervisor'] }
        },
        {
            path: '/admin/stages/:id/questions',
            name: 'admin-stage-questions',
            component: AdminStageQuestionsView,
            meta: { requiresAuth: true, roles: ['Admin', 'Supervisor'] }
        },
        {
            path: '/mentor',
            name: 'mentor',
            component: MentorView,
            meta: { requiresAuth: true, roles: ['Admin', 'Supervisor', 'Mentor'] }
        },
        {
            path: '/groups',
            name: 'groups',
            component: GroupsView,
            meta: { requiresAuth: true, roles: ['Admin', 'Supervisor', 'Mentor'] }
        },
        {
            path: '/groups/:id',
            name: 'group-details',
            component: GroupDetailsView,
            meta: { requiresAuth: true, roles: ['Admin', 'Supervisor', 'Mentor'] }
        }
    ]
})

router.beforeEach((to, from, next) => {
    const token = localStorage.getItem('token')
    const userStr = localStorage.getItem('user')

    if (to.meta.requiresAuth) {
        if (!token || !userStr) {
            next('/login')
            return
        }

        const user = JSON.parse(userStr)
        if (to.meta.roles && !to.meta.roles.includes(user.role)) {
            next('/welcome') // Redirect unauthorized
            return
        }

        // Auto-redirect students from /welcome to /student-home
        if (to.path === '/welcome' && user.role === 'Student') {
            next('/student-home')
            return
        }
    }

    // Prevent logged in user from going back to login
    if (to.path === '/login' && token) {
        const user = JSON.parse(localStorage.getItem('user') || '{}')
        if (user.role === 'Student') {
            next('/student-home')
        } else {
            next('/welcome')
        }
        return
    }

    next()
})

export default router
