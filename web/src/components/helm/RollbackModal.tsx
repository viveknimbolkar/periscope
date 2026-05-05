import { useHelmDiff } from "../../hooks/useHelm";
import { InlineDiff } from "../detail/yaml/InlineDiff";
import { ErrorState, LoadingState } from "../table/states";
import { ConfirmActionModal } from "../ui/ConfirmActionModal";

export function RollbackModal({
    open,
    onClose,
    onConfirm,
    releaseName,
    namespace,
    targetRevision,
    cluster,
    currentRevision
}: {
    open: boolean;
    onClose: () => void;
    onConfirm: () => void;
    releaseName: string;
    namespace: string;
    targetRevision: number;
    currentRevision: number;
    cluster: string;
}) {
    const diffQuery = useHelmDiff(
        cluster,
        namespace,
        releaseName,
        currentRevision,
        targetRevision
    );

    let bodyContent;
    if (diffQuery.isLoading) bodyContent = <LoadingState resource="diff" />;
    else if (diffQuery.isError) bodyContent = <ErrorState title="couldn't compute diff" message={(diffQuery.error as Error).message} />;
    else if (diffQuery.data) {
        bodyContent = (
            <div className="h-64 mt-4 overflow-y-auto border border-border">
                <InlineDiff original={diffQuery.data.from.yaml} proposed={diffQuery.data.to.yaml} />
            </div>
        );
    }
    return (
        <ConfirmActionModal
            open={open}
            title={`Are you sure you want to rollback ${releaseName} from revision ${currentRevision} → revision ${targetRevision}?`}
            onCancel={onClose}
            onConfirm={onConfirm}
            body={bodyContent}
            size="lg"
        />
    )
}   