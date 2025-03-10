import axios, { type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from "axios";

const service = axios.create({
    baseURL: 'http://127.0.0.1:8080/api',
    timeout: 10000,
})

export interface ApiResponse<T> {
    code: number,
    msg: string,
    data: T;
}

service.interceptors.request.use(
    (config: AxiosRequestConfig) => {
        config.headers = {
            'Content-Type': 'application/json',
            ...config.headers,
        }

        return config as InternalAxiosRequestConfig
    }
)

service.interceptors.response.use(
    (response: AxiosResponse) => {
        return response.data
    }
)

export default service