// Temporary type declarations for react-router-dom
declare module 'react-router-dom' {
  export function useNavigate(): (path: string) => void;
  export function useParams<T = Record<string, string>>(): T;
  export function BrowserRouter(props: { children: React.ReactNode }): JSX.Element;
  export function Routes(props: { children: React.ReactNode }): JSX.Element;
  export function Route(props: { path: string; element: React.ReactNode }): JSX.Element;
  export function Link(props: { to: string; children: React.ReactNode }): JSX.Element;
  export function Outlet(): JSX.Element;
}

// Temporary type declaration for React namespace
declare namespace React {
  namespace JSX {
    interface IntrinsicElements {
      [elemName: string]: any;
    }
  }
  interface ComponentType<P = {}> {
    (props: P): React.ReactElement | null;
  }
  interface ReactElement {
    type: any;
    props: any;
    key: string | null;
  }
  interface ReactNode {
    children?: React.ReactNode;
  }
}
