package grpcapi

import (
	"context"
	"database/sql"
	"log"
	"mailinglist/mdb"
	pb "mailinglist/proto"
	"net"
	"time"
  "google.golang.org/grpc"
)

type MailServer struct {
  pb.UnimplementedMailingListServiceServer
  db *sql.DB
}

func pbEntryToMdbEntry(pbEntry *pb.EmailEntry) mdb.EmailEntry  {
  t := time.Unix(pbEntry.ConfirmedAt, 0)
  return mdb.EmailEntry{
    Id: pbEntry.Id,
    Email: pbEntry.Email,
    ConfirmAt: &t,
    OptOut: pbEntry.OptOut,
  }
}

func MdbEntryToPbEntry(mdbEntry *mdb.EmailEntry) pb.EmailEntry  {
  return pb.EmailEntry{
    Id: mdbEntry.Id,
    Email: mdbEntry.Email,
    ConfirmedAt: mdbEntry.ConfirmAt.Unix(),
    OptOut: mdbEntry.OptOut,
  }
}

func emailResponse(db *sql.DB,email string)(*pb.EmailResponse,error)  {
  entry,err := mdb.GetEmail(db, email)
  if err != nil {
    return &pb.EmailResponse{}, err
  }
  if entry == nil {
    return &pb.EmailResponse{},nil
  }
  res := MdbEntryToPbEntry(entry)
  return &pb.EmailResponse{EmailEntry: &res},nil
  
}

func (s *MailServer) GetEmail(ctx context.Context , req *pb.GetEmailRequest) (*pb.EmailResponse,error)  {
  log.Printf("GRPC Get Email: %v \n", req)
  return emailResponse(s.db, req.EmailAddr)
  
}

func (s *MailServer) GetEmailBatch(ctx context.Context , req *pb.GetEmailBatchRequest) (*pb.GetEmailBatchResponse,error)  {
  log.Printf("GRPC Get Email batch: %v \n", req)
  params := mdb.GetEmailBatchQueryParams{
    Page: int(req.Page),
    Count: int(req.Count),
  }
  mdbEntries,err := mdb.GetEmailBatch(s.db, params) 
  if err != nil {
    return &pb.GetEmailBatchResponse{},err
  }
  pbEntries := make([]*pb.EmailEntry,0, len(mdbEntries))
  for i := 0; i < len(mdbEntries); i++ {
    entry := MdbEntryToPbEntry(&mdbEntries[i])
    pbEntries = append(pbEntries, &entry)
  }

  return &pb.GetEmailBatchResponse{EmailEntries: pbEntries},nil
  
}

func (s *MailServer) CreateEmail(ctx context.Context , req *pb.CreateEmailRequest) (*pb.EmailResponse,error)  {
  log.Printf("GRPC Create Email: %v \n", req)
  err := mdb.CreateEmail(s.db, req.EmailAddr)
  if err != nil {
    return &pb.EmailResponse{},err
  }
  return emailResponse(s.db, req.EmailAddr)
  
}
func (s *MailServer) UpdateEmail(ctx context.Context , req *pb.UpdateEmailRequest) (*pb.EmailResponse,error)  {
  log.Printf("GRPC Update Email: %v \n", req)
  mdbEntry := pbEntryToMdbEntry(req.EmailEntry)
  err := mdb.UpdateEmail(s.db, mdbEntry)
  if err != nil {
    return &pb.EmailResponse{},err
  }
  return emailResponse(s.db, mdbEntry.Email)
  
}
func (s *MailServer) DeleteEmail(ctx context.Context , req *pb.DeleteEmailRequest) (*pb.EmailResponse,error)  {
  log.Printf("GRPC Delete Email: %v \n", req)
  err := mdb.DeleteEmail(s.db, req.EmailAddr)
  if err != nil {
    return &pb.EmailResponse{},err
  }
  return emailResponse(s.db, req.EmailAddr)
  
}

func Serve(db *sql.DB,bind string)  {
  listener,err := net.Listen("tcp", bind)
  if err != nil {
    log.Fatalf("gRPC server error: failure to bind %v\n",bind)
  }
  grpcserver := grpc.NewServer()

  mailServer := MailServer{db:db}
  pb.RegisterMailingListServiceServer(grpcserver, &mailServer)
  log.Printf("gRPC API server listening on %v\n", bind)
  if err := grpcserver.Serve(listener);err != nil {
    log.Fatalf("gRPC server error: %v \n", err)
    
  }
  
}
